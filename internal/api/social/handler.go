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

	"xermess/internal/api/audit"
	"xermess/internal/api/respond"
	"xermess/internal/jose"
	"xermess/internal/model"
	"xermess/internal/store"
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
		Enabled:            true,
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

	if err := req.applyTo(provider, h.sealer, false); err != nil {
		respond.Failure(c, h.log, err, "validating a social provider failed")
		return
	}

	if err := h.store.SaveSocialProvider(c.Request.Context(), provider); err != nil {
		respond.Failure(c, h.log, err, "updating social provider failed")
		return
	}

	h.audit.RecordWith(c, "social_provider.updated", targetType, provider.ID.String(), map[string]any{
		"provider": provider.Name, "enabled": provider.Enabled,
	})

	h.answer(c, http.StatusOK, provider)
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
