// Package api builds the two HTTP servers and the table of which handler
// answers which path.
//
// The public server is the OAuth 2.0 / OpenID Connect provider and the account
// API, meant to face the internet. The admin server is the admin API the
// console calls, meant to be reachable only where administrators work. A path
// exists on only one of them.
//
// Handlers live one package per subject, each with handler.go, request.go,
// response.go and validation.go.
package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"loginer/internal/api/account"
	"loginer/internal/api/activity"
	"loginer/internal/api/adminroles"
	"loginer/internal/api/admins"
	"loginer/internal/api/apis"
	"loginer/internal/api/applications"
	"loginer/internal/api/audit"
	apiauth "loginer/internal/api/auth"
	"loginer/internal/api/caching"
	"loginer/internal/api/cors"
	"loginer/internal/api/csrf"
	"loginer/internal/api/fields"
	"loginer/internal/api/flows"
	"loginer/internal/api/keys"
	"loginer/internal/api/languages"
	adminmail "loginer/internal/api/mail"
	"loginer/internal/api/mfa"
	"loginer/internal/api/middleware"
	"loginer/internal/api/oauth"
	"loginer/internal/api/organization"
	"loginer/internal/api/otp"
	"loginer/internal/api/ratelimit"
	"loginer/internal/api/reference"
	"loginer/internal/api/respond"
	"loginer/internal/api/roles"
	"loginer/internal/api/session"
	"loginer/internal/api/sessions"
	"loginer/internal/api/setup"
	"loginer/internal/api/social"
	"loginer/internal/api/sso"
	"loginer/internal/api/users"
	"loginer/internal/auth"
	"loginer/internal/brand"
	"loginer/internal/cache"
	"loginer/internal/config"
	"loginer/internal/jose"
	"loginer/internal/model"
	"loginer/internal/oidc"
	"loginer/internal/store"
)

// tokenLimitMultiple loosens the provider's endpoints relative to the sign-in
// pages: one application backend may refresh tokens for a whole company from
// one address.
const tokenLimitMultiple = 10

// NewPublic builds the public server: provider, account API and health check.
// `shared` is the Redis rate limits count in, or nil for memory.
func NewPublic(cfg config.Config, log *slog.Logger, provider *oidc.Service, shared *cache.Redis) (*gin.Engine, error) {
	r, err := engine(cfg, log)
	if err != nil {
		return nil, err
	}

	document, err := reference.NewPublic(cfg.Issuer)
	if err != nil {
		return nil, err
	}

	registerPublicRoutes(r, publicHandlers{
		oauth:     oauth.New(provider, log, cfg.SecureUserCookies),
		account:   account.New(provider, log, cfg.SecureUserCookies),
		reference: document,
		limit:     ratelimit.New(cfg.RateLimit).Shared(shared.SessionDB(), "public").Middleware(),
		tokens:    ratelimit.New(cfg.RateLimit*tokenLimitMultiple).Shared(shared.SessionDB(), "token").Middleware(),
		csrf:      csrf.New(allowed(cfg.AccountURL, cfg.CORSOrigins)),
	})

	return r, nil
}

// NewAdmin builds the admin server: admin API and health check. It shares the
// provider with the public server, so key rotations apply to both.
func NewAdmin(cfg config.Config, st *store.Store, log *slog.Logger, provider *oidc.Service, shared *cache.Redis) (*gin.Engine, error) {
	r, err := engine(cfg, log)
	if err != nil {
		return nil, err
	}

	sealer, err := jose.NewSealer(cfg.SecretKey)
	if err != nil {
		return nil, err
	}

	document, err := reference.NewAdmin(config.Origin(cfg.AdminURL))
	if err != nil {
		return nil, err
	}

	service := auth.New(st, sealer, log, adminIssuer(cfg.AdminURL))
	recorder := audit.New(st, log)

	registerAdminRoutes(r, service, provider, log, adminHandlers{
		auth:         apiauth.New(service, st, recorder, log, cfg.SecureAdminCookies),
		mfa:          mfa.New(service, log),
		setup:        setup.New(st, log),
		users:        users.New(st, recorder, log),
		sessions:     sessions.New(st, recorder, log),
		fields:       fields.New(st, recorder, log),
		roles:        roles.New(st, recorder, log),
		admins:       admins.New(st, service, recorder, log),
		adminRoles:   adminroles.New(st, recorder, log),
		applications: applications.New(st, recorder, log, cfg.Issuer),
		organization: organization.New(st, recorder, log),
		social:       social.New(st, sealer, recorder, log, cfg.Issuer),
		sso:          sso.New(st, sealer, provider, recorder, log, cfg.Issuer),
		flows:        flows.New(st, recorder, log),
		languages:    languages.New(st, recorder, log),
		mail:         adminmail.New(st, sealer, recorder, log),
		otp:          otp.New(st, recorder, log),
		apis:         apis.New(st, recorder, log, cfg.Issuer),
		activity:     activity.New(st, log),
		keys:         keys.New(provider, recorder, log),
		caching:      caching.New(shared, recorder, log),
		reference:    document,
		limit:        ratelimit.New(cfg.RateLimit).Shared(shared.SessionDB(), "admin").Middleware(),
		csrf:         csrf.New(allowed(cfg.AdminURL, cfg.CORSOrigins)),
	})

	return r, nil
}

// adminIssuer is the authenticator app label: the panel's host, so staff with
// several installations can tell them apart.
func adminIssuer(adminURL string) string {
	if host := strings.TrimPrefix(strings.TrimPrefix(config.Origin(adminURL), "https://"), "http://"); host != "" {
		return brand.Name + " (" + host + ")"
	}
	return brand.Name
}

// engine is what both servers start from. gin.New has no middleware, so the
// slog logger is the only one.
func engine(cfg config.Config, log *slog.Logger) (*gin.Engine, error) {
	r := gin.New()

	// Gin believes every X-Forwarded-For header unless told otherwise, which
	// would let any caller choose the address the log records for it.
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("trusted proxies: %w", err)
	}

	// Which origins may call the API is configuration, so the handler is
	// built here and handed to the chain that puts it in order.
	r.Use(middleware.Chain(log, cors.New(cfg.CORSOrigins))...)

	return r, nil
}

// allowed is the origins that may change things through a server's API: the
// origin of the app it serves, and any extra CORS origin configured.
func allowed(appURL string, extra []string) []string {
	return append([]string{config.Origin(appURL)}, extra...)
}

// publicHandlers and adminHandlers are one of each, so the route tables below
// read as tables.
type publicHandlers struct {
	oauth     *oauth.Handler
	account   *account.Handler
	reference *reference.Handler
	limit     gin.HandlerFunc
	// tokens is the limit on the provider's own endpoints, which is not the
	// one the sign-in pages are held to; see tokenLimitMultiple.
	tokens gin.HandlerFunc
	csrf   gin.HandlerFunc
}

type adminHandlers struct {
	auth         *apiauth.Handler
	mfa          *mfa.Handler
	setup        *setup.Handler
	users        *users.Handler
	sessions     *sessions.Handler
	fields       *fields.Handler
	roles        *roles.Handler
	admins       *admins.Handler
	adminRoles   *adminroles.Handler
	applications *applications.Handler
	apis         *apis.Handler
	organization *organization.Handler
	social       *social.Handler
	sso          *sso.Handler
	flows        *flows.Handler
	languages    *languages.Handler
	mail         *adminmail.Handler
	otp          *otp.Handler
	activity     *activity.Handler
	keys         *keys.Handler
	caching      *caching.Handler
	reference    *reference.Handler
	limit        gin.HandlerFunc
	csrf         gin.HandlerFunc
}

// registerPublicRoutes mounts every public route. What a path does is in the
// handler, but that a path exists — and on which server — is only ever here.
func registerPublicRoutes(r *gin.Engine, h publicHandlers) {
	r.GET("/healthz", health)

	// The OAuth 2.0 / OpenID Connect provider, at paths fixed by discovery
	// rather than versioned.
	r.GET(oidc.PathDiscovery, h.oauth.Discovery)
	r.GET(oidc.PathJWKS, h.oauth.JWKS)
	// Authorize writes a row per request, so it shares the token endpoints'
	// loose per-address limit.
	r.GET(oidc.PathAuthorize, h.tokens, h.oauth.Authorize)
	r.POST(oidc.PathAuthorize, h.tokens, h.oauth.Authorize)
	// Endpoints taking client credentials or a token have their own loose
	// per-address limit (tokenLimitMultiple).
	r.POST(oidc.PathToken, h.tokens, h.oauth.Token)
	r.GET(oidc.PathUserInfo, h.oauth.UserInfo)
	r.POST(oidc.PathUserInfo, h.oauth.UserInfo)
	r.GET(oidc.PathLogout, h.oauth.Logout)
	r.POST(oidc.PathLogout, h.oauth.Logout)
	r.POST(oidc.PathRevoke, h.tokens, h.oauth.Revoke)
	r.POST(oidc.PathIntrospect, h.tokens, h.oauth.Introspect)

	// This server's OpenAPI document, beside discovery so the same proxy routes
	// reach it.
	r.GET("/.well-known/openapi.json", h.reference.Document)

	// Social sign-in. The callback also answers POST and sits outside CSRF
	// because Apple posts its answer from its own form.
	r.GET(oidc.PathSocialStart, h.limit, h.oauth.SocialStart)
	r.GET(oidc.PathSocialCallback, h.oauth.SocialCallback)
	r.POST(oidc.PathSocialCallback, h.oauth.SocialCallback)

	// Enterprise SSO. The SAML ACS takes the provider's posted form, so it is
	// outside CSRF; start and ACS are rate limited because anyone can call
	// them.
	r.GET(oidc.PathSSOStart, h.limit, h.oauth.SSOStart)
	r.GET(oidc.PathSSOCallback, h.oauth.SSOCallback)
	r.POST(oidc.PathSSOACS, h.limit, h.oauth.SSOAssertion)
	r.GET(oidc.PathSSOMetadata, h.oauth.SSOMetadata)

	// Everything under /api/v1 is called by the id app with the user's
	// session cookie, so it only takes changes from the id app's origin.
	v1 := r.Group("/api/v1", h.csrf)
	{
		// The id app's API. Signing in needs no session; password and email
		// endpoints are rate limited per address.
		accounts := v1.Group("/account")
		accounts.GET("/organization", h.account.Organization)
		accounts.GET("/social-providers", h.account.SocialProviders)
		accounts.GET("/sso", h.account.SSOButtons)
		accounts.POST("/sso/discover", h.limit, h.account.DiscoverSSO)
		accounts.GET("/requests/:handle", h.account.Request)
		accounts.GET("/login-options", h.account.LoginOptions)
		accounts.GET("/languages", h.account.Languages)
		accounts.GET("/languages/:code", h.account.LanguageText)
		accounts.GET("/applications/:client_id", h.account.Application)
		accounts.POST("/login", h.limit, h.account.Login)
		// Emailed code sign-in: both rate limited (code guesses, and sending
		// mail).
		accounts.POST("/login/code", h.limit, h.account.Code)
		accounts.POST("/login/code/resend", h.limit, h.account.ResendCode)
		accounts.POST("/register", h.limit, h.account.Register)
		accounts.POST("/forgot-password", h.limit, h.account.ForgotPassword)
		accounts.GET("/reset-password", h.account.CheckReset)
		accounts.POST("/reset-password", h.limit, h.account.ResetPassword)
		accounts.POST("/verify-email", h.limit, h.account.VerifyEmail)
		accounts.POST("/logout", h.account.Logout)

		// A user managing their own account: only ever the one whose session
		// the request carries.
		own := accounts.Group("", h.account.RequireSession)
		own.GET("/me", h.account.Me)
		own.PATCH("/me", h.account.UpdateMe)
		own.POST("/password", h.limit, h.account.ChangePassword)
		// Moving to another sign-in address sends mail to whatever is typed,
		// so it is limited per address like the rest that do.
		own.POST("/email", h.limit, h.account.ChangeEmail)
		own.GET("/sessions", h.account.Sessions)
		own.DELETE("/sessions/:id", h.account.EndSession)
		own.GET("/connected-applications", h.account.Applications)
		own.DELETE("/connected-applications/:client_id", h.account.Disconnect)
	}

	r.NoRoute(notFound)
}

// registerAdminRoutes mounts every admin route. None of them is on the public
// server.
func registerAdminRoutes(r *gin.Engine, service *auth.Service, provider *oidc.Service, log *slog.Logger, h adminHandlers) {
	r.GET("/healthz", health)

	// Everything here is called by the console with the administrator's session
	// cookie, so it only takes changes from the console's origin.
	v1 := r.Group("/api/v1", h.csrf)
	{
		// Setup and sign-in cannot require a session. Setup is refused once any
		// administrator exists.
		v1.GET("/admin/setup", h.setup.Status)
		v1.POST("/admin/setup", h.limit, h.setup.Create)
		v1.POST("/admin/auth/login", h.limit, h.auth.Login)

		// The rest of admin sign-in: the session's state, finishing with a
		// code, and signing out at any step.
		v1.GET("/admin/auth/session", h.auth.State)
		v1.POST("/admin/auth/mfa", h.limit, h.auth.VerifyMFA)
		v1.POST("/admin/auth/logout", h.auth.Logout)

		// Setting up an authenticator: for a signed-in administrator, and for
		// one who has to before they may do anything else.
		setup := v1.Group("/admin/mfa", session.RequireSetup(service))
		setup.GET("", h.mfa.Status)
		setup.POST("/totp", h.limit, h.mfa.Begin)
		setup.POST("/totp/confirm", h.limit, h.mfa.Confirm)

		signedIn := v1.Group("/admin", session.Require(service))
		{
			// Everything about the caller's own account is open to every
			// administrator who can sign in.
			signedIn.GET("/me", h.auth.Me)
			signedIn.PATCH("/me", h.limit, h.auth.UpdateMe)
			signedIn.POST("/me/password", h.limit, h.auth.ChangePassword)

			// Changing a second factor that is on takes a code from it.
			signedIn.DELETE("/mfa/totp", h.limit, h.mfa.Disable)
			signedIn.POST("/mfa/recovery-codes", h.limit, h.mfa.RecoveryCodes)
			signedIn.GET("/sessions", h.auth.Sessions)
			signedIn.DELETE("/sessions/:id", h.auth.EndSession)
			signedIn.DELETE("/sessions", h.auth.EndOtherSessions)

			// The admin API as OpenAPI. Unlike the public one it is not for
			// strangers: which routes exist is the panel's business.
			signedIn.GET("/openapi.json", h.reference.Document)

			// Everything below is guarded by the administrator's permissions;
			// the panel hiding a control is not the check. These routes also
			// accept an admin API access token whose scopes stand in for roles;
			// routes about the signed-in person, and super-admin routes, stay
			// session-only.
			managed := v1.Group("/admin", session.RequireAny(service, provider, log))

			activity := managed.Group("", session.Can(model.PermActivityRead))
			activity.GET("/overview", h.activity.Overview)
			activity.GET("/logs", h.activity.Logs)
			activity.GET("/logs/export", h.activity.Export)

			// Users and their fields are shared by every application, so these
			// are panel-wide permissions.
			readUsers := managed.Group("", session.Can(model.PermUsersRead))
			readUsers.GET("/users", h.users.List)
			readUsers.GET("/users/:id", h.users.Get)
			readUsers.GET("/users/:id/roles", h.users.Roles)
			readUsers.GET("/users/:id/role-mappings", h.users.RoleMappings)
			readUsers.GET("/user-fields", h.fields.List)
			readUsers.GET("/user-sessions", h.sessions.List)

			writeUsers := managed.Group("", session.Can(model.PermUsersWrite))
			writeUsers.POST("/users", h.users.Create)
			writeUsers.PATCH("/users/:id", h.users.Update)
			writeUsers.DELETE("/users/:id", h.users.Delete)
			writeUsers.DELETE("/users/:id/social-accounts/:identity", h.users.Disconnect)
			writeUsers.DELETE("/users/:id/sessions", h.sessions.SignOutUser)
			writeUsers.DELETE("/user-sessions/:id", h.sessions.End)

			writeFields := managed.Group("", session.Can(model.PermUserFieldsWrite))
			writeFields.POST("/user-fields", h.fields.Create)
			writeFields.PATCH("/user-fields/:id", h.fields.Update)
			writeFields.DELETE("/user-fields/:id", h.fields.Delete)

			// The organisation's single settings record.
			managed.GET("/organization", session.Can(model.PermOrganizationRead), h.organization.Get)
			signedIn.PATCH("/organization", session.Can(model.PermOrganizationWrite), h.organization.Update)

			// Social providers; changing them decides which outside accounts
			// get in, so it is its own permission.
			readSocial := managed.Group("", session.Can(model.PermSocialRead))
			readSocial.GET("/social-providers", h.social.List)
			readSocial.GET("/social-providers/:id", h.social.Get)

			// Reading a secret back takes the permission that could replace
			// it, and is recorded like a change.
			writeSocial := managed.Group("", session.Can(model.PermSocialWrite))
			writeSocial.GET("/social-providers/:id/secret", h.social.Secret)
			writeSocial.POST("/social-providers", h.social.Create)
			writeSocial.PATCH("/social-providers/:id", h.social.Update)
			writeSocial.DELETE("/social-providers/:id", h.social.Delete)

			// SSO connections; changing or testing them decides who signs in
			// with which roles, so it is its own permission.
			readSSO := managed.Group("", session.Can(model.PermSSORead))
			readSSO.GET("/sso-connections", h.sso.List)
			readSSO.GET("/sso-connections/:id", h.sso.Get)

			writeSSO := managed.Group("", session.Can(model.PermSSOWrite))
			writeSSO.POST("/sso-connections", h.sso.Create)
			writeSSO.POST("/sso-connections/test", h.sso.Test)
			writeSSO.PATCH("/sso-connections/:id", h.sso.Update)
			writeSSO.DELETE("/sso-connections/:id", h.sso.Delete)
			writeSSO.POST("/sso-connections/:id/refresh-metadata", h.sso.RefreshMetadata)

			// Login flows; writing one decides what sign-in asks for, so it is
			// its own permission.
			readFlows := managed.Group("", session.Can(model.PermLoginFlowsRead))
			readFlows.GET("/login-flows", h.flows.List)
			readFlows.GET("/login-flows/:id", h.flows.Get)

			writeFlows := managed.Group("", session.Can(model.PermLoginFlowsWrite))
			writeFlows.POST("/login-flows", h.flows.Create)
			writeFlows.PATCH("/login-flows/:id", h.flows.Update)
			writeFlows.DELETE("/login-flows/:id", h.flows.Delete)

			// Languages and their text; changing them changes every sign-in
			// page, so it needs languages.write.
			readLanguages := managed.Group("", session.Can(model.PermLanguagesRead))
			readLanguages.GET("/languages", h.languages.List)
			readLanguages.GET("/languages/:code/translations/:app", h.languages.Translation)

			writeLanguages := managed.Group("", session.Can(model.PermLanguagesWrite))
			writeLanguages.POST("/languages", h.languages.Create)
			writeLanguages.PATCH("/languages/:code", h.languages.Update)
			writeLanguages.DELETE("/languages/:code", h.languages.Delete)
			writeLanguages.PUT("/languages/:code/translations/:app", h.languages.SaveTranslation)

			// Applications and their roles. Routes only check the administrator
			// reaches some application; each handler checks the one requested.
			apps := managed.Group("", session.CanAnywhere(model.PermApplicationsRead))
			apps.GET("/applications", h.applications.List)
			apps.GET("/applications/:id", h.applications.Get)
			apps.PATCH("/applications/:id", h.applications.Update)
			apps.POST("/applications/:id/secret", h.applications.RotateSecret)

			// User roles, global and per application. Handlers narrow to the
			// scopes the administrator reaches and check writes against the
			// role's scope.
			roles := managed.Group("", session.CanAnywhere(model.PermUsersRead, model.PermApplicationsRead))
			roles.GET("/user-roles", h.roles.List)
			roles.GET("/user-roles/:id", h.roles.Get)

			writeRoles := managed.Group("", session.CanAnywhere(model.PermUserRolesWrite))
			writeRoles.POST("/user-roles", h.roles.Create)
			writeRoles.PATCH("/user-roles/:id", h.roles.Update)
			writeRoles.DELETE("/user-roles/:id", h.roles.Delete)

			// Keycloak's role mapping: giving a user roles and taking them
			// away, each role checked against its own scope.
			assign := managed.Group("", session.CanAnywhere(model.PermRoleAssignmentsWrite))
			assign.POST("/users/:id/role-mappings", h.users.AssignRoles)
			assign.DELETE("/users/:id/role-mappings/:role", h.users.UnassignRole)

			// What an application may do with each API, and a preview of the
			// tokens it would get. Each handler checks the application.
			apps.GET("/applications/:id/apis", h.applications.APIAccess)
			apps.PUT("/applications/:id/apis/:api", h.applications.AuthorizeAPI)
			apps.DELETE("/applications/:id/apis/:api", h.applications.RevokeAPI)
			apps.POST("/applications/:id/token-preview", h.applications.TokenPreview)

			// APIs. Readable by anyone who can see them or configure an
			// application's access; changing them is panel-wide.
			readAPIs := managed.Group("", session.CanAnywhere(model.PermAPIsRead, model.PermApplicationsWrite))
			readAPIs.GET("/apis", h.apis.List)
			readAPIs.GET("/apis/:id", h.apis.Get)
			readAPIs.GET("/apis/:id/applications", h.apis.Applications)

			// An API's log names every application, visible or not, so it needs
			// apis.read itself.
			managed.GET("/apis/:id/logs", session.Can(model.PermAPIsRead), h.apis.Logs)

			writeAPIs := managed.Group("", session.Can(model.PermAPIsWrite))
			writeAPIs.POST("/apis", h.apis.Create)
			writeAPIs.PATCH("/apis/:id", h.apis.Update)
			writeAPIs.DELETE("/apis/:id", h.apis.Delete)

			// Registering and removing applications changes what exists at
			// all, so it takes applications.write for the whole panel.
			registerApps := managed.Group("", session.Can(model.PermApplicationsWrite))
			registerApps.POST("/applications", h.applications.Create)
			registerApps.DELETE("/applications/:id", h.applications.Delete)

			// The administrators themselves, and their roles: a super
			// admin's alone, whatever other roles grant.
			super := signedIn.Group("", session.RequireSuperAdmin())
			// How administrators are made to sign in, which is the panel's
			// setting rather than the configuration's after the first start.
			super.GET("/security", h.admins.Security)
			super.PATCH("/security", h.admins.UpdateSecurity)

			super.GET("/admins", h.admins.List)
			super.POST("/admins", h.admins.Create)
			super.GET("/admins/:id", h.admins.Get)
			super.PATCH("/admins/:id", h.admins.Update)
			super.DELETE("/admins/:id", h.admins.Delete)
			super.DELETE("/admins/:id/mfa", h.admins.ResetMFA)

			// Mail settings carry the SMTP password and the email text decides
			// what lands in inboxes, so both are super-admin only. The test
			// send is rate limited.
			super.GET("/mail", h.mail.Get)
			super.PATCH("/mail", h.mail.Update)
			super.POST("/mail/test", h.limit, h.mail.Test)
			super.GET("/mail/content", h.mail.Content)
			super.PUT("/mail/content/:code", h.mail.SaveContent)

			// Emailed one-time code settings decide how hard sign-in is, so
			// they are super-admin only.
			super.GET("/otp", h.otp.Get)
			super.PATCH("/otp", h.otp.Update)

			// The keys tokens are signed with. Rotating them decides which
			// tokens every API trusts, so it is a super admin's alone.
			super.GET("/signing-keys", h.keys.List)
			super.POST("/signing-keys/rotate", h.limit, h.keys.Rotate)

			// Redis inspection and clearing. A hand-written value is what pages
			// then show, so it is super-admin only.
			super.GET("/cache", h.caching.Overview)
			super.GET("/cache/:database/keys", h.caching.Keys)
			super.GET("/cache/:database/key", h.caching.Key)
			super.PUT("/cache/:database/key", h.caching.UpdateKey)
			super.DELETE("/cache/:database/key", h.caching.DeleteKey)
			super.POST("/cache/:database/groups/:group/clear", h.caching.ClearGroup)
			super.DELETE("/cache/:database", h.caching.Flush)

			super.GET("/admin-permissions", h.adminRoles.Permissions)
			super.GET("/admin-roles", h.adminRoles.List)
			super.POST("/admin-roles", h.adminRoles.Create)
			super.PATCH("/admin-roles/:id", h.adminRoles.Update)
			super.DELETE("/admin-roles/:id", h.adminRoles.Delete)
		}
	}

	r.NoRoute(notFound)
}

// health says the server is up.
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// notFound answers any path that no route matched.
func notFound(c *gin.Context) {
	respond.Fail(c, respond.NotFoundAny)
}
