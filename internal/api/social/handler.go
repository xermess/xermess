// Package social answers the endpoints that configure signing in with an
// account somewhere else: the providers an administrator registers this
// server with, and what each of them is allowed to do here.
//
// Nothing in this package signs anybody in — that is internal/oidc, which
// reads these records. This is the panel's half: what is stored, and what the
// panel is told about it, which is never a secret it has stored.
package social

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/internal/api/audit"
	"loginer/internal/api/respond"
	"loginer/internal/jose"
	"loginer/internal/model"
	"loginer/internal/store"
)

// Handler holds what these endpoints need.
type Handler struct {
	store  *store.Store
	sealer *jose.Sealer
	audit  audit.Recorder
	log    *slog.Logger
	// issuer is what the callback addresses are built from, so the panel can
	// show an administrator what to register with the provider.
	issuer string
}

// New returns a Handler.
func New(st *store.Store, sealer *jose.Sealer, recorder audit.Recorder, log *slog.Logger, issuer string) *Handler {
	return &Handler{store: st, sealer: sealer, audit: recorder, log: log, issuer: issuer}
}

// List returns every configured provider, and the kinds a new one may be.
func (h *Handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	providers, err := h.store.SocialProviders(ctx, false)
	if err != nil {
		respond.Failure(c, h.log, err, "listing social providers failed")
		return
	}

	counts, err := h.store.SocialIdentityCounts(ctx)
	if err != nil {
		respond.Failure(c, h.log, err, "counting social identities failed")
		return
	}

	c.JSON(http.StatusOK, newListResponse(providers, counts, h.issuer))
}

// Get returns one provider.
func (h *Handler) Get(c *gin.Context) {
	provider, ok := h.find(c)
	if !ok {
		return
	}

	h.answer(c, http.StatusOK, provider)
}

// Create registers a provider. It is not reachable by anyone signing in until
// somebody has signed in with it once — the credentials are only ever proved
// right by the provider itself.
func (h *Handler) Create(c *gin.Context) {
	var req providerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	provider := &model.SocialProvider{
		// What a new provider is like before anything is typed in: enabled,
		// linking addresses the provider has checked, and taking new accounts.
		IsEnabled:          true,
		LinkVerifiedEmails: true,
		AllowRegistration:  true,
	}

	if err := req.applyTo(provider, h.sealer, true); err != nil {
		respond.Failure(c, h.log, err, "validating a social provider failed")
		return
	}

	if req.Position == nil {
		provider.Position = h.store.NextSocialPosition(c.Request.Context())
	}

	if err := h.store.CreateSocialProvider(c.Request.Context(), provider); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			respond.Conflict(c, "a provider with that identifier already exists")
			return
		}

		respond.Failure(c, h.log, err, "creating social provider failed")
		return
	}

	h.audit.RecordWith(c, "social_provider.created", targetType, provider.ID.String(), map[string]any{
		"provider": provider.Name, "kind": string(provider.Kind),
	})

	h.answer(c, http.StatusCreated, provider)
}

// Update changes a provider's settings. Its kind and its identifier are not
// among them: both are in the address registered with the provider, and a
// sign-in that comes back to an address nobody answers is a worse failure
// than having to register a second provider.
func (h *Handler) Update(c *gin.Context) {
	provider, ok := h.find(c)
	if !ok {
		return
	}

	var req providerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	before := identityNamespace(provider)

	if err := req.applyTo(provider, h.sealer, false); err != nil {
		respond.Failure(c, h.log, err, "validating a social provider failed")
		return
	}

	// Pointing a provider at a different upstream — new endpoints, a new
	// client, a new Apple team — changes whose subjects its stored identities
	// name. A stranger at the new upstream whose subject matched would sign
	// in as the old owner. SSO forgets its identities in this case; the social
	// identities of most kinds cannot re-link (their address is never proven),
	// so here the change is refused while any identity still hangs on it, and
	// the administrator registers a new provider instead.
	if identityNamespace(provider) != before {
		counts, err := h.store.SocialIdentityCounts(c.Request.Context())
		if err != nil {
			respond.Failure(c, h.log, err, "counting social identities failed")
			return
		}
		if counts[provider.ID] > 0 {
			respond.Fail(c, providerRepointed)
			return
		}
	}

	if err := h.store.SaveSocialProvider(c.Request.Context(), provider); err != nil {
		respond.Failure(c, h.log, err, "updating social provider failed")
		return
	}

	h.audit.RecordWith(c, "social_provider.updated", targetType, provider.ID.String(), map[string]any{
		"provider": provider.Name, "is_enabled": provider.IsEnabled,
	})

	h.answer(c, http.StatusOK, provider)
}

// providerRepointed is the answer to changing where a provider with linked
// identities signs people in from.
var providerRepointed = respond.Define(http.StatusConflict, "social_provider_repointed", respond.Admin)

// identityNamespace is what decides whether two settings of a provider name
// the same people: the part of a provider a subject is only unique within. A
// change to it means a stored subject now belongs to a different person.
//
// For a custom kind that is the upstream itself — its token and userinfo
// endpoints and the client. Apple scopes its subject to the team; Facebook to
// the app. Google, Yandex and VK keep a subject stable across the app
// registration, so rotating their client id strands nobody.
func identityNamespace(p *model.SocialProvider) string {
	_, token, userInfo := p.Endpoints()

	switch p.Kind {
	case model.SocialOAuth2, model.SocialOIDC:
		return "custom\x00" + token + "\x00" + userInfo + "\x00" + p.ClientID
	case model.SocialApple:
		return "apple\x00" + p.TeamID
	case model.SocialFacebook:
		return "facebook\x00" + p.ClientID
	default:
		return ""
	}
}

// Secret answers with the client secret itself, for an administrator who has
// to check what is configured against the provider's console.
//
// It is stored encrypted rather than hashed, so it can be read back — which
// makes reading it an event worth recording. It takes the permission that
// could replace it anyway, and every reading is written to the log with who
// asked.
func (h *Handler) Secret(c *gin.Context) {
	provider, ok := h.find(c)
	if !ok {
		return
	}

	sealed, what := provider.ClientSecret, "client_secret"
	if provider.Spec().SignedSecret {
		sealed, what = provider.PrivateKey, "private_key"
	}

	if len(sealed) == 0 {
		respond.NotFound(c, "there is no secret stored for this provider")
		return
	}

	secret, err := h.sealer.OpenBytes(sealed)
	if err != nil {
		respond.Failure(c, h.log, err, "opening a social provider secret failed")
		return
	}

	h.audit.RecordWith(c, "social_provider.secret_read", targetType, provider.ID.String(), map[string]any{
		"provider": provider.Name, "secret": what,
	})

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"secret": string(secret), "kind": what})
}

// Delete removes a provider, and every identity held at it. The accounts stay
// — a user who had no password keeps one way in fewer, which is why the panel
// says how many people that is before it asks.
func (h *Handler) Delete(c *gin.Context) {
	provider, ok := h.find(c)
	if !ok {
		return
	}

	if err := h.store.DeleteSocialProvider(c.Request.Context(), provider); err != nil {
		respond.Failure(c, h.log, err, "deleting social provider failed")
		return
	}

	h.audit.RecordWith(c, "social_provider.deleted", targetType, provider.ID.String(), map[string]any{
		"provider": provider.Name,
	})

	c.Status(http.StatusNoContent)
}

// find loads the provider named in the path, answering the request itself if
// there is no such provider.
func (h *Handler) find(c *gin.Context) (*model.SocialProvider, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond.BadRequest(c, "that is not a provider id")
		return nil, false
	}

	provider, err := h.store.SocialProvider(c.Request.Context(), id)

	switch {
	case errors.Is(err, store.ErrNotFound):
		respond.NotFound(c, "no such provider")
		return nil, false
	case err != nil:
		respond.Failure(c, h.log, err, "loading social provider failed")
		return nil, false
	}

	return provider, true
}

// answer writes one provider, with how many users hold an identity at it.
func (h *Handler) answer(c *gin.Context, status int, provider *model.SocialProvider) {
	counts, err := h.store.SocialIdentityCounts(c.Request.Context())
	if err != nil {
		respond.Failure(c, h.log, err, "counting social identities failed")
		return
	}

	c.JSON(status, gin.H{"provider": newProviderResponse(*provider, h.issuer, counts[provider.ID])})
}
