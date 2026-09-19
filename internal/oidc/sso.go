package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"xermess/internal/model"
	"xermess/internal/store"
	"xermess/locales"
)

// Enterprise single sign-on: an organisation's own identity provider signing
// its people in, over OpenID Connect or SAML 2.0 (model.SSOConnection).
//
// As with social sign-in this server is the client — the relying party, the
// service provider — and the paths carry the connection's slug. The callback,
// the assertion consumer service and the metadata are given to the provider,
// so they may not move once anybody signs in through them.
//
// What is checked, because the provider's word is what signs somebody in:
//
//   - OpenID Connect: the code is bound to this sign-in by PKCE and the state,
//     and the id_token's signature is verified against the provider's
//     published keys, along with its issuer, audience, expiry and nonce;
//   - SAML: the response has to be signed by the certificate in the provider's
//     metadata, be meant for this service provider, be in time, and answer the
//     request this server sent — an unsolicited, IdP-initiated response is
//     refused, since nothing ties it to the browser presenting it;
//   - both: the address has to be at one of the connection's own domains.
const (
	PathSSOStart    = "/oauth2/sso/:slug/start"
	PathSSOCallback = "/oauth2/sso/:slug/callback"
	PathSSOACS      = "/oauth2/sso/:slug/acs"
	PathSSOMetadata = "/oauth2/sso/:slug/metadata"
)

// SSOLoginLifetime is how long a sign-in may take at the provider.
const SSOLoginLifetime = 15 * time.Minute

// The refusals of signing in through a connection, as the sign-in pages say
// them.
var (
	// ErrSSOUnknown is a connection that is not there, or is off.
	ErrSSOUnknown = problem("sso_unknown", "the connection is not available")
	// ErrSSOExpired is a sign-in that took too long, came back twice, or was
	// never started here — which is also what an IdP-initiated one is.
	ErrSSOExpired = problem("sso_expired", "the single sign-on expired or was used")
	// ErrSSOUpstream is the provider refusing, failing, or answering with
	// something that does not verify. What it was is logged, and written to
	// the activity log against the connection (ssoFailed).
	ErrSSOUpstream = problem("sso_upstream", "the identity provider did not complete the sign-in")
	// ErrSSONoEmail is a provider that said nothing this server can sign in:
	// no address, or one it says it has not verified.
	ErrSSONoEmail = problem("sso_no_email", "the identity provider gave no usable address")
	// ErrSSODomainMismatch is an address at a domain the connection does not
	// own.
	ErrSSODomainMismatch = problem("sso_domain_mismatch", "the address is outside the connection's domains")
	// ErrSSOLinkRefused is an address that already has an account, at a
	// connection set to refuse rather than link.
	ErrSSOLinkRefused = problem("sso_link_refused", "the address has an account the connection does not link")
	// ErrSSONoAccount is somebody new, at a connection that does not make
	// accounts.
	ErrSSONoAccount = problem("sso_no_account", "the connection does not create accounts")
	// ErrSSORequired is a password sign-in, registration or reset for an
	// address whose domain has to sign in through its provider. Returned as
	// an *SSORequired, which says which.
	ErrSSORequired = problem("sso_required", "the address has to sign in through single sign-on")
)

// SSORequired is ErrSSORequired with the connection the address belongs to,
// so the sign-in page can send the person there.
type SSORequired struct {
	SSOButton
}

func (e *SSORequired) Error() string { return ErrSSORequired.Error() }
func (e *SSORequired) Unwrap() error { return ErrSSORequired }

// SSOButton is a connection as a sign-in page offers it.
type SSOButton struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
}

func ssoButton(connection *model.SSOConnection) SSOButton {
	return SSOButton{Slug: connection.Slug, Name: connection.Name, Protocol: string(connection.Protocol)}
}

// SSOButtons are the connections the sign-in page shows a button for, and
// whether any connection signs people in at all — which is whether the page
// offers "Sign in with SSO" for the ones without a button.
func (s *Service) SSOButtons(ctx context.Context) ([]SSOButton, bool, error) {
	connections, err := s.store.SSOButtons(ctx)
	if err != nil {
		return nil, false, err
	}

	available, err := s.store.AnySSOConnectionEnabled(ctx)
	if err != nil {
		return nil, false, err
	}

	out := make([]SSOButton, 0, len(connections))
	for i := range connections {
		out = append(out, ssoButton(&connections[i]))
	}

	return out, available, nil
}

// DiscoverSSO is the connection an address signs in through, for the sign-in
// page's "Sign in with SSO": the enabled connection owning its domain, whether
// it is enforced, or store.ErrNotFound.
//
// It says which domains have a connection, which is not a secret — the page
// would send anybody who typed such an address there anyway.
func (s *Service) DiscoverSSO(ctx context.Context, email string) (SSOButton, bool, error) {
	connection, err := s.store.SSOConnectionForEmail(ctx, email)
	if err != nil {
		return SSOButton{}, false, err
	}

	return ssoButton(connection), connection.EnforceDomains, nil
}

// ssoRequiredFor refuses the password ways in — signing in, registering,
// resetting — for an address whose domain has to use its connection.
func (s *Service) ssoRequiredFor(ctx context.Context, email string) error {
	connection, err := s.store.SSOConnectionForEmail(ctx, email)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil
	case err != nil:
		return err
	case !connection.EnforceDomains:
		return nil
	}

	return &SSORequired{ssoButton(connection)}
}

// ssoConnection is a connection by slug, if people may sign in through it.
func (s *Service) ssoConnection(ctx context.Context, slug string) (*model.SSOConnection, error) {
	connection, err := s.store.SSOConnectionBySlug(ctx, slug)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, ErrSSOUnknown
	case err != nil:
		return nil, err
	case !connection.Enabled:
		return nil, ErrSSOUnknown
	}

	return connection, nil
}

// StartSSO is where to send the browser to sign in through a connection.
// `loginHint` is the address the person typed, passed on so they are not
// asked for it again.
func (s *Service) StartSSO(ctx context.Context, slug, request, next, loginHint string) (string, error) {
	connection, err := s.ssoConnection(ctx, slug)
	if err != nil {
		return "", err
	}

	state, stateHash, err := model.NewSecret()
	if err != nil {
		return "", err
	}

	login := model.SSOLogin{
		StateHash:    stateHash,
		ConnectionID: connection.ID,
		Request:      request,
		Next:         next,
		ExpiresAt:    s.now().Add(SSOLoginLifetime),
	}

	var location string

	switch connection.Protocol {
	case model.SSOProtocolOIDC:
		location, err = s.startOIDC(ctx, connection, &login, state, loginHint)
	case model.SSOProtocolSAML:
		location, err = s.startSAML(connection, &login, state)
	default:
		err = ErrSSOUnknown
	}
	if err != nil {
		return "", s.ssoFailed(ctx, connection, Client{}, err)
	}

	if err := s.store.CreateSSOLogin(ctx, &login); err != nil {
		return "", err
	}

	return location, nil
}

// ssoPerson is who a provider said somebody is.
type ssoPerson struct {
	Subject   string
	Email     string
	FirstName string
	LastName  string
	Groups    []string

	// Sent are the names of the claims or attributes the provider sent, so a
	// refusal can say what there was instead of what was looked for.
	Sent []string
}

// CompleteSSOCallback turns the code an OpenID Connect provider sent the
// browser back with into a session.
func (s *Service) CompleteSSOCallback(ctx context.Context, slug, code, state string, client Client) (*SocialResult, error) {
	connection, login, err := s.takeSSOLogin(ctx, slug, state)
	if err != nil {
		return nil, err
	}
	if connection.Protocol != model.SSOProtocolOIDC || code == "" {
		return nil, ErrSSOExpired
	}

	person, err := s.completeOIDC(ctx, connection, login, code)
	if err != nil {
		return nil, s.ssoFailed(ctx, connection, client, err)
	}

	return s.finishSSO(ctx, connection, login, person, client)
}

// CompleteSSOAssertion turns the response a SAML provider posted to the
// assertion consumer service into a session.
func (s *Service) CompleteSSOAssertion(ctx context.Context, slug string, r *http.Request, client Client) (*SocialResult, error) {
	if err := r.ParseForm(); err != nil {
		return nil, ErrSSOExpired
	}

	// No relay state is an unsolicited response: refused, whatever it says.
	connection, login, err := s.takeSSOLogin(ctx, slug, r.PostForm.Get("RelayState"))
	if err != nil {
		return nil, err
	}
	if connection.Protocol != model.SSOProtocolSAML {
		return nil, ErrSSOExpired
	}

	person, err := s.completeSAML(connection, login, r)
	if err != nil {
		return nil, s.ssoFailed(ctx, connection, client, err)
	}

	return s.finishSSO(ctx, connection, login, person, client)
}

// RefuseSSO is a provider sending the browser back with an OAuth error rather
// than a code — the person cancelled, or the provider would not let the
// connection sign them in: an unknown scope, a redirect URI or a client it
// does not know. The sign-in is used up, and the refusal recorded against the
// connection, but only for a state this server issued, so a link anybody can
// make does not write to the activity log.
func (s *Service) RefuseSSO(ctx context.Context, slug, state, code, description string, client Client) error {
	connection, _, err := s.takeSSOLogin(ctx, slug, state)
	if err != nil {
		return err
	}

	reason := code
	if description != "" {
		reason += ": " + description
	}

	return s.ssoFailed(ctx, connection, client, upstream("authorize", errors.New(reason)))
}

// ssoRefusal is a sign-in through a connection that did not complete: what
// the person signing in is shown, and — for the administrator, in the
// activity log — at which step, and exactly why.
type ssoRefusal struct {
	shown *Problem
	step  string
	err   error
}

// upstream is the provider failing, or answering with something that does
// not verify.
func upstream(step string, err error) error {
	return &ssoRefusal{shown: ErrSSOUpstream, step: step, err: err}
}

// refused is the provider answering, but with somebody this server will not
// sign in: shown as the reason, with the detail for the log.
func refused(shown *Problem, format string, args ...any) error {
	return &ssoRefusal{shown: shown, step: "claims", err: fmt.Errorf(format, args...)}
}

func (f *ssoRefusal) Error() string { return f.step + ": " + f.err.Error() }
func (f *ssoRefusal) Unwrap() error { return f.err }

// ssoFailedAction is what the activity log calls a sign-in the provider did
// not complete.
const ssoFailedAction = "sso_connection.sign_in_failed"

// ssoFailed turns a refusal into the problem the person is shown, and writes
// what it was to the activity log against the connection. The person signing
// in is told nothing a provider said, which is no business of theirs; the
// administrator setting the connection up needs all of it, and should not
// have to go looking in the server's log.
func (s *Service) ssoFailed(ctx context.Context, connection *model.SSOConnection, client Client, err error) error {
	var failure *ssoRefusal
	if !errors.As(err, &failure) {
		return err
	}

	s.log.Error("an identity provider did not complete a sign-in",
		"connection", connection.Slug, "step", failure.step, "error", failure.err)

	reason := failure.err.Error()
	if len(reason) > 500 {
		reason = reason[:500] + "…"
	}

	entry := model.AuditLog{
		Action:     ssoFailedAction,
		TargetType: "sso_connection",
		TargetID:   connection.ID.String(),
		IP:         client.IP,
		UserAgent:  client.UserAgent,
		Metadata: map[string]any{
			"connection": connection.Name,
			"step":       failure.step,
			"shown":      failure.shown.Code,
			"reason":     reason,
		},
	}
	if err := s.store.WriteAudit(ctx, &entry); err != nil {
		s.log.Error("writing the activity log failed", "error", err, "action", ssoFailedAction)
	}

	return failure.shown
}

// takeSSOLogin is the sign-in a state belongs to, used up, and the enabled
// connection it was for.
func (s *Service) takeSSOLogin(ctx context.Context, slug, state string) (*model.SSOConnection, *model.SSOLogin, error) {
	connection, err := s.ssoConnection(ctx, slug)
	if err != nil {
		return nil, nil, err
	}

	if state == "" {
		return nil, nil, ErrSSOExpired
	}

	login, err := s.store.TakeSSOLogin(ctx, model.HashSecret(state), s.now())
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, nil, ErrSSOExpired
	case err != nil:
		return nil, nil, err
	case login.ConnectionID != connection.ID:
		// A state made for another connection is not an answer from this one.
		return nil, nil, ErrSSOExpired
	}

	return connection, login, nil
}

// finishSSO signs the person in and says where to send them.
func (s *Service) finishSSO(
	ctx context.Context,
	connection *model.SSOConnection,
	login *model.SSOLogin,
	person *ssoPerson,
	client Client,
) (*SocialResult, error) {
	result, err := s.signInThroughSSO(ctx, connection, person, client)
	if err != nil {
		return nil, s.ssoFailed(ctx, connection, client, err)
	}

	return &SocialResult{SignIn: result, Request: login.Request, Next: login.Next, Provider: connection.Name}, nil
}

// signInThroughSSO is the account a verified person signs in to — theirs
// already, one their address has, or a new one — brought up to date with
// what the provider says.
func (s *Service) signInThroughSSO(
	ctx context.Context,
	connection *model.SSOConnection,
	person *ssoPerson,
	client Client,
) (*SignInResult, error) {
	now := s.now()
	email := strings.ToLower(strings.TrimSpace(person.Email))

	switch {
	case person.Subject == "":
		return nil, refused(ErrSSONoEmail, "the provider named nobody: no subject")
	case email == "":
		return nil, refused(ErrSSONoEmail, "no address in %s; the provider sent %s",
			strings.Join(connection.Attribute("email"), " or "), strings.Join(person.Sent, ", "))
	case !connection.OwnsEmail(email):
		return nil, refused(ErrSSODomainMismatch, "%s is outside the connection's domains", email)
	}

	var user *model.User

	identity, err := s.store.SSOIdentity(ctx, connection.ID, person.Subject)
	switch {
	case err == nil:
		if user, err = s.store.User(ctx, identity.UserID); err != nil {
			return nil, err
		}
		if err := s.store.MarkSSOIdentityUsed(ctx, identity, email, now); err != nil {
			return nil, err
		}
	case !errors.Is(err, store.ErrNotFound):
		return nil, err
	default:
		if user, err = s.ssoAccountFor(ctx, connection, person, email, client); err != nil {
			return nil, err
		}

		identity := &model.SSOIdentity{
			UserID: user.ID, ConnectionID: connection.ID, Subject: person.Subject,
			Email: email, LastLoginAt: &now,
		}
		if err := s.store.CreateSSOIdentity(ctx, identity); err != nil {
			return nil, err
		}
	}

	if !user.CanSignIn(now) {
		s.record(ctx, user, user.Email, "user.login_blocked", client, map[string]any{
			"reason": "inactive", "connection": connection.Name,
		})
		return nil, ErrInvalidCredentials
	}

	if err := s.syncFromSSO(ctx, connection, user, person, client); err != nil {
		return nil, err
	}

	result, err := s.startSession(ctx, user, client, "user.login")
	if err != nil {
		return nil, err
	}

	s.log.Info("a user signed in through single sign-on", "connection", connection.Slug, "user", user.ID)

	return result, nil
}

// ssoAccountFor is the account somebody signing in through a connection for
// the first time gets: the one their address already has, if the connection
// links, or a new one, if it makes them.
func (s *Service) ssoAccountFor(
	ctx context.Context,
	connection *model.SSOConnection,
	person *ssoPerson,
	email string,
	client Client,
) (*model.User, error) {
	user, err := s.store.UserByEmail(ctx, email)
	switch {
	case err == nil:
		if connection.Matching != model.SSOMatchLink {
			return nil, ErrSSOLinkRefused
		}
		// Somebody can register an address they do not own and wait for its
		// owner to arrive through the provider, keeping the password they
		// set; an account that never proved its address is not linked.
		if !user.EmailVerified {
			return nil, refused(ErrSSOLinkRefused, "the account for %s never verified its address, so it is not linked", email)
		}

		s.record(ctx, user, user.Email, "user.identity_connected", client, map[string]any{
			"provider": connection.Name,
		})

		return user, nil
	case !errors.Is(err, store.ErrNotFound):
		return nil, err
	}

	if !connection.CreateUsers {
		return nil, ErrSSONoAccount
	}

	// The provider owns the domain, so the address is as verified as it gets.
	user = &model.User{
		Email:         email,
		EmailVerified: true,
		FirstName:     truncate(person.FirstName, 100),
		LastName:      truncate(person.LastName, 100),
		IsActive:      true,
		Data:          map[string]any{},
	}

	roles, err := s.store.DefaultUserRoles(ctx)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	if err := s.store.CreateUser(ctx, user); errors.Is(err, store.ErrDuplicate) {
		// Made between the two queries above: the next sign-in links it.
		return nil, ErrSSOExpired
	} else if err != nil {
		return nil, err
	}

	s.record(ctx, user, user.Email, "user.registered", client, map[string]any{"provider": connection.Name})

	return user, nil
}

// syncFromSSO keeps the account in step with the provider: its name, when the
// connection says the provider is where it is kept, and the roles the
// provider's groups give.
func (s *Service) syncFromSSO(ctx context.Context, connection *model.SSOConnection, user *model.User, person *ssoPerson, client Client) error {
	if connection.SyncProfile {
		first, last := truncate(person.FirstName, 100), truncate(person.LastName, 100)
		if (first != "" && first != user.FirstName) || (last != "" && last != user.LastName) {
			if first != "" {
				user.FirstName = first
			}
			if last != "" {
				user.LastName = last
			}
			if err := s.store.SaveUser(ctx, user); err != nil {
				return err
			}
		}
	}

	if len(connection.RoleMappings) == 0 {
		return nil
	}

	held := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		held = append(held, role.ID.String())
	}

	given := connection.RoleMappings.For(person.Groups)

	var add []model.UserRole
	for _, id := range given {
		if !slices.Contains(held, id.String()) {
			add = append(add, model.UserRole{Base: model.Base{ID: id}})
		}
	}

	var removed []string
	if connection.SyncRoles {
		for _, id := range connection.RoleMappings.Roles() {
			if slices.Contains(held, id.String()) && !slices.Contains(given, id) {
				if err := s.store.RemoveUserRole(ctx, user.ID, id); err != nil {
					return err
				}
				removed = append(removed, id.String())
			}
		}
	}

	if len(add) > 0 {
		// A mapping can name a role since deleted; the ones that exist are
		// given, and the rest are nothing to give.
		existing, err := s.store.UserRolesByID(ctx, rolesIDs(add))
		if err != nil {
			return err
		}
		if err := s.store.AddUserRoles(ctx, user.ID, existing); err != nil {
			return err
		}
	}

	if len(add) > 0 || len(removed) > 0 {
		s.record(ctx, user, user.Email, "user.roles_synced", client, map[string]any{
			"connection": connection.Name, "added": len(add), "removed": len(removed),
		})
	}

	return nil
}

// ---- the error page -------------------------------------------------------

// SSOErrorPage is the sign-in app's error page for a failed single sign-on,
// naming the reason the way the API does so the page says it in the reader's
// language.
func (s *Service) SSOErrorPage(err error) string {
	reason := ErrSSOUpstream

	var known *Problem
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		reason = ErrSocialBlocked
	case errors.As(err, &known) && strings.HasPrefix(known.Code, "sso_"):
		reason = known
	}

	description, ok := locales.Text(locales.ID, "error."+reason.Code)
	if !ok {
		description = reason.message
	}

	return withQuery(s.accountURL+PageError, url.Values{
		"error":             {SSOErrorCode},
		"reason":            {reason.Code},
		"error_description": {description},
	})
}

// SSOErrorCode is what the error page is told a failed single sign-on was.
const SSOErrorCode = "sso_sign_in_failed"

// IsSSOFailure reports whether an error is one of the refusals a person is
// shown, rather than the server's own failure.
func IsSSOFailure(err error) bool {
	var known *Problem
	return errors.Is(err, ErrInvalidCredentials) || errors.As(err, &known) && strings.HasPrefix(known.Code, "sso_")
}

func rolesIDs(roles []model.UserRole) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(roles))
	for _, role := range roles {
		out = append(out, role.ID)
	}

	return out
}
