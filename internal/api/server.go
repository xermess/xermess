// Package api builds the HTTP servers: the middleware every request passes
// through, and the tables of which handler answers which path.
//
// There are two servers, on two listeners, and a path exists on only one of
// them. The public server is the OAuth 2.0 / OpenID Connect provider and the
// account API the id app calls: it is meant to face the internet. The admin
// server is the admin API the console calls, and nothing else: it is meant to be
// reachable only from where administrators work. Keeping them apart is what
// makes "the admin panel is internal" true of the API too, rather than only of
// the page that calls it.
//
// The handlers themselves live one directory down, one package per subject:
// auth signs administrators in, users manages user records, fields the
// columns those records are made of, applications the OAuth clients that sign
// users in, roles the roles users hold in each application, admins and
// adminroles the administrators and what they may do, organization the
// settings of the installation itself, activity reports on what has happened,
// oauth and account are the provider and the users' own API. Each of those
// packages is the same four files — the handler, the requests it accepts, the
// answers it gives, and the rules it holds them to — so finding your way
// around a new one is the same as finding your way around the last.
package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"xermess/internal/api/account"
	"xermess/internal/api/activity"
	"xermess/internal/api/adminroles"
	"xermess/internal/api/admins"
	"xermess/internal/api/apis"
	"xermess/internal/api/applications"
	"xermess/internal/api/audit"
	apiauth "xermess/internal/api/auth"
	"xermess/internal/api/cors"
	"xermess/internal/api/csrf"
	"xermess/internal/api/database"
	"xermess/internal/api/fields"
	"xermess/internal/api/flows"
	"xermess/internal/api/keys"
	"xermess/internal/api/languages"
	"xermess/internal/api/mfa"
	"xermess/internal/api/middleware"
	"xermess/internal/api/oauth"
	"xermess/internal/api/organization"
	"xermess/internal/api/ratelimit"
	"xermess/internal/api/respond"
	"xermess/internal/api/roles"
	"xermess/internal/api/session"
	"xermess/internal/api/setup"
	"xermess/internal/api/social"
	"xermess/internal/api/sso"
	"xermess/internal/api/users"
	"xermess/internal/auth"
	"xermess/internal/cache"
	"xermess/internal/config"
	"xermess/internal/jose"
	"xermess/internal/model"
	"xermess/internal/oidc"
	"xermess/internal/store"
)

// NewPublic builds the public server: the provider, the account API, and a
// health check. `provider` is built by the caller because building it reads
// the signing keys from the database. `shared` is the Redis the rate limit
// counts in, or nil to count in memory.
func NewPublic(cfg config.Config, log *slog.Logger, provider *oidc.Service, shared *cache.Cache) (*gin.Engine, error) {
	r, err := engine(cfg, log)
	if err != nil {
		return nil, err
	}

	registerPublicRoutes(r, publicHandlers{
		oauth:   oauth.New(provider, log, cfg.SecureUserCookies),
		account: account.New(provider, log, cfg.SecureUserCookies),
		limit:   ratelimit.New(cfg.RateLimit).Shared(shared, "public").Middleware(),
		csrf:    csrf.New(allowed(cfg.AccountURL, cfg.CORSOrigins)),
	})

	return r, nil
}

// NewAdmin builds the admin server: the admin API, and a health check.
// `provider` is the same provider the public server answers with, so rotating
// its signing keys here takes effect there at once.
func NewAdmin(cfg config.Config, st *store.Store, log *slog.Logger, provider *oidc.Service, shared *cache.Cache) (*gin.Engine, error) {
	r, err := engine(cfg, log)
	if err != nil {
		return nil, err
	}

	sealer, err := jose.NewSealer(cfg.SecretKey)
	if err != nil {
		return nil, err
	}

	service := auth.New(st, sealer, log, adminIssuer(cfg.AdminURL))
	recorder := audit.New(st, log)

	registerAdminRoutes(r, service, adminHandlers{
		auth:         apiauth.New(service, st, log, cfg.SecureAdminCookies),
		mfa:          mfa.New(service, log),
		setup:        setup.New(st, log),
		users:        users.New(st, recorder, log),
		fields:       fields.New(st, recorder, log),
		roles:        roles.New(st, recorder, log),
		admins:       admins.New(st, service, recorder, log),
		adminRoles:   adminroles.New(st, recorder, log),
		applications: applications.New(st, recorder, log, cfg.Issuer),
		organization: organization.New(st, recorder, log),
		social:       social.New(st, sealer, recorder, log, cfg.Issuer),
		sso:          sso.New(st, sealer, provider, recorder, log, cfg.Issuer),
		flows:        flows.New(st, recorder, log),
		database:     database.New(st, log),
		languages:    languages.New(st, recorder, log),
		apis:         apis.New(st, recorder, log, cfg.Issuer),
		activity:     activity.New(st, log),
		keys:         keys.New(provider, recorder, log),
		limit:        ratelimit.New(cfg.RateLimit).Shared(shared, "admin").Middleware(),
		csrf:         csrf.New(allowed(cfg.AdminURL, cfg.CORSOrigins)),
	})

	return r, nil
}

// adminIssuer is what authenticator apps list an administrator's account
// under: the panel's host, so staff with several installations can tell them
// apart.
func adminIssuer(adminURL string) string {
	if host := strings.TrimPrefix(strings.TrimPrefix(config.Origin(adminURL), "https://"), "http://"); host != "" {
		return "xermess (" + host + ")"
	}
	return "xermess admin"
}

// engine is what both servers start from: middleware first, in the order
// every request passes through them.
//
// gin.New starts with no middleware, unlike gin.Default, which adds Gin's own
// logger. We want the slog one instead, so the whole server logs the same way.
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
	oauth   *oauth.Handler
	account *account.Handler
	limit   gin.HandlerFunc
	csrf    gin.HandlerFunc
}

type adminHandlers struct {
	auth         *apiauth.Handler
	mfa          *mfa.Handler
	setup        *setup.Handler
	users        *users.Handler
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
	database     *database.Handler
	languages    *languages.Handler
	activity     *activity.Handler
	keys         *keys.Handler
	limit        gin.HandlerFunc
	csrf         gin.HandlerFunc
}

// registerPublicRoutes mounts every public route. What a path does is in the
// handler, but that a path exists — and on which server — is only ever here.
func registerPublicRoutes(r *gin.Engine, h publicHandlers) {
	r.GET("/healthz", health)

	// The OAuth 2.0 and OpenID Connect provider. These paths are fixed by the
	// discovery document rather than versioned with the API: every client
	// library finds them from there.
	r.GET(oidc.PathDiscovery, h.oauth.Discovery)
	r.GET(oidc.PathJWKS, h.oauth.JWKS)
	r.GET(oidc.PathAuthorize, h.oauth.Authorize)
	r.POST(oidc.PathAuthorize, h.oauth.Authorize)
	r.POST(oidc.PathToken, h.oauth.Token)
	r.GET(oidc.PathUserInfo, h.oauth.UserInfo)
	r.POST(oidc.PathUserInfo, h.oauth.UserInfo)
	r.GET(oidc.PathLogout, h.oauth.Logout)
	r.POST(oidc.PathLogout, h.oauth.Logout)
	r.POST(oidc.PathRevoke, h.oauth.Revoke)
	r.POST(oidc.PathIntrospect, h.oauth.Introspect)

	// Signing in with an account somewhere else. The callback answers POST
	// as well, because Apple posts its answer rather than redirecting with
	// it; it is outside the CSRF group for the same reason — the form comes
	// from Apple, not from this server's own app.
	r.GET(oidc.PathSocialStart, h.oauth.SocialStart)
	r.GET(oidc.PathSocialCallback, h.oauth.SocialCallback)
	r.POST(oidc.PathSocialCallback, h.oauth.SocialCallback)

	// Signing in through an organisation's own identity provider. The
	// assertion consumer service takes a SAML provider's posted response, and
	// is outside the CSRF group for the reason Apple's callback is: the form
	// comes from the provider. The metadata is what the provider is set up
	// from.
	// Starting writes a sign-in and the ACS verifies an XML signature, both
	// for anybody, so both are rate limited per address.
	r.GET(oidc.PathSSOStart, h.limit, h.oauth.SSOStart)
	r.GET(oidc.PathSSOCallback, h.oauth.SSOCallback)
	r.POST(oidc.PathSSOACS, h.limit, h.oauth.SSOAssertion)
	r.GET(oidc.PathSSOMetadata, h.oauth.SSOMetadata)

	// Everything under /api/v1 is called by the id app with the user's
	// session cookie, so it only takes changes from the id app's origin.
	v1 := r.Group("/api/v1", h.csrf)
	{
		// What the id app calls. Signing in needs no session: it is how a
		// user gets one, and what a page may do is decided by the sign-in
		// handle it was given, and by the password. The routes that take a
		// password or send an email are rate limited per address.
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
		accounts.POST("/register", h.limit, h.account.Register)
		accounts.POST("/forgot-password", h.limit, h.account.ForgotPassword)
		accounts.GET("/reset-password", h.account.CheckReset)
		accounts.POST("/reset-password", h.limit, h.account.ResetPassword)
		accounts.POST("/logout", h.account.Logout)

		// A user managing their own account: only ever the one whose session
		// the request carries.
		own := accounts.Group("", h.account.RequireSession)
		own.GET("/me", h.account.Me)
		own.PATCH("/me", h.account.UpdateMe)
		own.POST("/password", h.limit, h.account.ChangePassword)
		own.GET("/sessions", h.account.Sessions)
		own.DELETE("/sessions/:id", h.account.EndSession)
		own.GET("/connected-applications", h.account.Applications)
		own.DELETE("/connected-applications/:client_id", h.account.Disconnect)
	}

	r.NoRoute(notFound)
}

// registerAdminRoutes mounts every admin route. None of them is on the public
// server.
func registerAdminRoutes(r *gin.Engine, service *auth.Service, h adminHandlers) {
	r.GET("/healthz", health)

	// Everything here is called by the console with the administrator's session
	// cookie, so it only takes changes from the console's origin.
	v1 := r.Group("/api/v1", h.csrf)
	{
		// Setting the panel up and signing in are the routes that cannot
		// require a session: before the first there is no account, and before
		// the second no way to prove one. Creating an administrator is
		// refused as soon as there is one, which is what keeps the first of
		// those from being a way in.
		v1.GET("/admin/setup", h.setup.Status)
		v1.POST("/admin/setup", h.limit, h.setup.Create)
		v1.POST("/admin/auth/login", h.limit, h.auth.Login)

		// The panel's own text, which the sign-in page is drawn in before
		// there is anybody to be signed in. It is the words on the page and
		// nothing more.
		v1.GET("/admin/panel/languages", h.languages.PanelLanguages)
		v1.GET("/admin/panel/languages/:code", h.languages.PanelText)

		// Signing in, the rest of the way. The state says which step a
		// session is at; a code finishes a sign-in waiting for one; signing
		// out works at any step.
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

			// Changing a second factor that is on takes a code from it.
			signedIn.DELETE("/mfa/totp", h.limit, h.mfa.Disable)
			signedIn.POST("/mfa/recovery-codes", h.limit, h.mfa.RecoveryCodes)
			signedIn.GET("/sessions", h.auth.Sessions)

			// The rest is guarded by what the administrator's roles allow.
			// The panel hides what someone cannot do; these checks are what
			// actually stops them.
			activity := signedIn.Group("", session.Can(model.PermActivityRead))
			activity.GET("/overview", h.activity.Overview)
			activity.GET("/logs", h.activity.Logs)

			// The users an organisation manages and the fields their records
			// are made of. Users are shared by every application, so these
			// are whole-panel permissions.
			readUsers := signedIn.Group("", session.Can(model.PermUsersRead))
			readUsers.GET("/users", h.users.List)
			readUsers.GET("/users/:id", h.users.Get)
			readUsers.GET("/users/:id/roles", h.users.Roles)
			readUsers.GET("/users/:id/role-mappings", h.users.RoleMappings)
			readUsers.GET("/user-fields", h.fields.List)

			writeUsers := signedIn.Group("", session.Can(model.PermUsersWrite))
			writeUsers.POST("/users", h.users.Create)
			writeUsers.PATCH("/users/:id", h.users.Update)
			writeUsers.DELETE("/users/:id", h.users.Delete)
			writeUsers.DELETE("/users/:id/social-accounts/:identity", h.users.Disconnect)

			writeFields := signedIn.Group("", session.Can(model.PermUserFieldsWrite))
			writeFields.POST("/user-fields", h.fields.Create)
			writeFields.PATCH("/user-fields/:id", h.fields.Update)
			writeFields.DELETE("/user-fields/:id", h.fields.Delete)

			// The organisation the installation belongs to: one record of
			// settings, read by anyone whose roles allow the page and
			// written by anyone allowed to change it.
			signedIn.GET("/organization", session.Can(model.PermOrganizationRead), h.organization.Get)
			signedIn.PATCH("/organization", session.Can(model.PermOrganizationWrite), h.organization.Update)

			// The providers users may sign in with. Registering one decides
			// which accounts elsewhere reach this server, so changing them is
			// its own permission, apart from reading them.
			readSocial := signedIn.Group("", session.Can(model.PermSocialRead))
			readSocial.GET("/social-providers", h.social.List)
			readSocial.GET("/social-providers/:id", h.social.Get)

			// Reading a secret back takes the permission that could replace
			// it, and is recorded like a change.
			writeSocial := signedIn.Group("", session.Can(model.PermSocialWrite))
			writeSocial.GET("/social-providers/:id/secret", h.social.Secret)
			writeSocial.POST("/social-providers", h.social.Create)
			writeSocial.PATCH("/social-providers/:id", h.social.Update)
			writeSocial.DELETE("/social-providers/:id", h.social.Delete)

			// The organisations' own identity providers. Connecting one decides
			// who may sign in, as whom, and with which roles, so changing them
			// is its own permission; trying a provider is part of setting it
			// up, so it takes the same.
			readSSO := signedIn.Group("", session.Can(model.PermSSORead))
			readSSO.GET("/sso-connections", h.sso.List)
			readSSO.GET("/sso-connections/:id", h.sso.Get)

			writeSSO := signedIn.Group("", session.Can(model.PermSSOWrite))
			writeSSO.POST("/sso-connections", h.sso.Create)
			writeSSO.POST("/sso-connections/test", h.sso.Test)
			writeSSO.PATCH("/sso-connections/:id", h.sso.Update)
			writeSSO.DELETE("/sso-connections/:id", h.sso.Delete)
			writeSSO.POST("/sso-connections/:id/refresh-metadata", h.sso.RefreshMetadata)

			// The login flows applications sign their users in with, and the
			// steps one can be made of. Writing a flow decides what a
			// sign-in asks for, so it is its own permission.
			readFlows := signedIn.Group("", session.Can(model.PermLoginFlowsRead))
			readFlows.GET("/login-flows", h.flows.List)
			readFlows.GET("/login-flows/:id", h.flows.Get)

			writeFlows := signedIn.Group("", session.Can(model.PermLoginFlowsWrite))
			writeFlows.POST("/login-flows", h.flows.Create)
			writeFlows.PATCH("/login-flows/:id", h.flows.Update)
			writeFlows.DELETE("/login-flows/:id", h.flows.Delete)

			// The languages, and their text for each app. Adding, rewording
			// and removing one changes what every sign-in page says, so it
			// takes languages.write; reading the text back takes only read.
			readLanguages := signedIn.Group("", session.Can(model.PermLanguagesRead))
			readLanguages.GET("/languages", h.languages.List)
			readLanguages.GET("/languages/:code/translations/:app", h.languages.Translation)

			writeLanguages := signedIn.Group("", session.Can(model.PermLanguagesWrite))
			writeLanguages.POST("/languages", h.languages.Create)
			writeLanguages.PATCH("/languages/:code", h.languages.Update)
			writeLanguages.DELETE("/languages/:code", h.languages.Delete)
			writeLanguages.PUT("/languages/:code/translations/:app", h.languages.SaveTranslation)

			// The server's own tables, read row by row. There is no endpoint
			// here that writes one: a record is changed on the page that
			// knows what it is.
			readDatabase := signedIn.Group("", session.Can(model.PermDatabaseRead))
			readDatabase.GET("/database/tables", h.database.Tables)
			readDatabase.GET("/database/tables/:table", h.database.Table)

			// Applications, the roles each defines, and who holds them. A
			// role can grant these for one application, so the routes only
			// check the administrator can reach some application; each
			// handler then checks the one the request is about.
			apps := signedIn.Group("", session.CanAnywhere(model.PermApplicationsRead))
			apps.GET("/applications", h.applications.List)
			apps.GET("/applications/:id", h.applications.Get)
			apps.PATCH("/applications/:id", h.applications.Update)
			apps.POST("/applications/:id/secret", h.applications.RotateSecret)

			// Roles: global ones, which belong to no application, and each
			// application's. Anyone who can see users or some application can
			// read them; each handler narrows to the scopes the administrator
			// reaches and checks writes against the role's scope.
			roles := signedIn.Group("", session.CanAnywhere(model.PermUsersRead, model.PermApplicationsRead))
			roles.GET("/user-roles", h.roles.List)
			roles.GET("/user-roles/:id", h.roles.Get)

			writeRoles := signedIn.Group("", session.CanAnywhere(model.PermUserRolesWrite))
			writeRoles.POST("/user-roles", h.roles.Create)
			writeRoles.PATCH("/user-roles/:id", h.roles.Update)
			writeRoles.DELETE("/user-roles/:id", h.roles.Delete)

			// Keycloak's role mapping: giving a user roles and taking them
			// away, each role checked against its own scope.
			assign := signedIn.Group("", session.CanAnywhere(model.PermRoleAssignmentsWrite))
			assign.POST("/users/:id/role-mappings", h.users.AssignRoles)
			assign.DELETE("/users/:id/role-mappings/:role", h.users.UnassignRole)

			// What an application may do with each API, and a preview of the
			// tokens it would get. Each handler checks the application.
			apps.GET("/applications/:id/apis", h.applications.APIAccess)
			apps.PUT("/applications/:id/apis/:api", h.applications.AuthorizeAPI)
			apps.DELETE("/applications/:id/apis/:api", h.applications.RevokeAPI)
			apps.POST("/applications/:id/token-preview", h.applications.TokenPreview)

			// APIs: the resource servers tokens are issued for. Anyone who can
			// see them or configure an application's access to them may read
			// them; changing them is a whole-panel permission.
			readAPIs := signedIn.Group("", session.CanAnywhere(model.PermAPIsRead, model.PermApplicationsWrite))
			readAPIs.GET("/apis", h.apis.List)
			readAPIs.GET("/apis/:id", h.apis.Get)
			readAPIs.GET("/apis/:id/applications", h.apis.Applications)

			// An API's log names every application given or refused access,
			// including ones an administrator of a few applications cannot
			// see, so it takes apis.read itself.
			signedIn.GET("/apis/:id/logs", session.Can(model.PermAPIsRead), h.apis.Logs)

			writeAPIs := signedIn.Group("", session.Can(model.PermAPIsWrite))
			writeAPIs.POST("/apis", h.apis.Create)
			writeAPIs.PATCH("/apis/:id", h.apis.Update)
			writeAPIs.DELETE("/apis/:id", h.apis.Delete)

			// Registering and removing applications changes what exists at
			// all, so it takes applications.write for the whole panel.
			registerApps := signedIn.Group("", session.Can(model.PermApplicationsWrite))
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

			// The keys tokens are signed with. Rotating them decides which
			// tokens every API trusts, so it is a super admin's alone.
			super.GET("/signing-keys", h.keys.List)
			super.POST("/signing-keys/rotate", h.limit, h.keys.Rotate)

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
