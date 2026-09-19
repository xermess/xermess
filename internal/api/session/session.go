// Package session is everything about who is signed in, over HTTP: the cookie
// the browser carries, the check every admin route passes through, and the
// way a handler asks who is making the request.
package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"xermess/internal/api/respond"
	"xermess/internal/auth"
	"xermess/internal/model"
)

// Cookie carries the session token between the browser and the server. It is
// HttpOnly so no script can read it, which is what keeps a cross-site
// scripting bug from turning into a stolen session.
const Cookie = "xermess_session"

// key is the Gin context key the signed-in administrator is stored under.
const key = "admin"

// Set stores the token in the browser. `secure` adds the Secure flag, which
// can only be on once the API is served over HTTPS.
func Set(c *gin.Context, token string, secure bool) {
	write(c, token, int(auth.SessionLifetime.Seconds()), secure)
}

// Clear removes it again.
func Clear(c *gin.Context, secure bool) {
	write(c, "", -1, secure)
}

func write(c *gin.Context, value string, maxAge int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     Cookie,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// Require refuses the request unless it carries a session for an
// administrator who may sign in. Handlers behind it can call Admin without
// checking.
func Require(service *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(Cookie)

		user, err := service.Authenticate(c.Request.Context(), token)
		if err != nil {
			respond.Abort(c, respond.NotSignedIn)
			return
		}

		c.Set(key, user)

		c.Next()
	}
}

// RequireSetup lets through a signed-in administrator, and one who is half
// signed in waiting to set up a second factor — and nobody else. It guards
// the endpoints that set a factor up, which is all such a session may do.
func RequireSetup(service *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(Cookie)

		user, _, state, err := service.Session(c.Request.Context(), token)
		if err != nil || (state != auth.StateSignedIn && state != auth.StateEnroll) {
			respond.Abort(c, respond.NotSignedIn)
			return
		}

		c.Set(key, user)

		c.Next()
	}
}

// Can refuses the request unless the signed-in administrator holds a role
// granting the permission. It goes after Require, which loads the roles.
func Can(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if admin := Admin(c); admin == nil || !admin.HasPermission(permission) {
			respond.Abort(c, respond.NotAllowed)
			return
		}

		c.Next()
	}
}

// CanAnywhere refuses the request unless the signed-in administrator holds one
// of the permissions for the whole panel or for at least one application. It
// guards the routes whose handlers then narrow what they do to the
// applications the administrator can reach, using Allowed and Reach.
func CanAnywhere(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := Admin(c)

		for _, permission := range permissions {
			if admin != nil && admin.HasPermissionAnywhere(permission) {
				c.Next()
				return
			}
		}

		respond.Abort(c, respond.NotAllowed)
	}
}

// Allowed reports whether the signed-in administrator holds the permission
// for one application: a whole-panel role grants it, or one scoped to that
// application does.
func Allowed(c *gin.Context, permission string, application uuid.UUID) bool {
	admin := Admin(c)

	return admin != nil && admin.HasPermissionFor(permission, &application)
}

// AllowedScope reports whether the signed-in administrator holds the
// permission for a role's scope: for its application, or for the whole panel
// when the role is global (a nil application).
func AllowedScope(c *gin.Context, permission string, application *uuid.UUID) bool {
	if application == nil {
		admin := Admin(c)
		return admin != nil && admin.HasPermission(permission)
	}

	return Allowed(c, permission, *application)
}

// SeesRole reports whether the signed-in administrator can see a user role:
// a global role is seen by everyone who reaches the endpoints that list
// roles, and an application's roles by whoever can read the application.
func SeesRole(c *gin.Context, role model.UserRole) bool {
	return role.Global() || Allowed(c, model.PermApplicationsRead, *role.ApplicationID)
}

// Reach says which applications the signed-in administrator holds the
// permission for, as a store filter: nil for every application, otherwise
// the ids of the ones they can reach, which may be none.
func Reach(c *gin.Context, permission string) []uuid.UUID {
	admin := Admin(c)
	if admin == nil {
		return []uuid.UUID{}
	}

	all, ids := admin.ApplicationsWith(permission)
	if all {
		return nil
	}

	return ids
}

// superAdminOnly is what an administrator who is not a super admin is told
// by the routes that are a super admin's alone.
var superAdminOnly = respond.Define(http.StatusForbidden, "super_admin_only", respond.Admin)

// RequireSuperAdmin refuses the request unless the signed-in administrator is
// a super admin: managing administrators and their roles is theirs alone.
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if admin := Admin(c); admin == nil || !admin.IsSuperAdmin() {
			respond.Abort(c, superAdminOnly)
			return
		}

		c.Next()
	}
}

// Admin returns the administrator making the request. It is only valid behind
// Require, which is the only thing that sets it.
func Admin(c *gin.Context) *model.AdminUser {
	user, _ := c.Get(key)
	admin, _ := user.(*model.AdminUser)

	return admin
}
