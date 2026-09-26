package applications

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/internal/api/query"
	"loginer/internal/model"
	"loginer/internal/store"
)

// targetType is what these records are called in the activity log.
const targetType = "application"

// maxPageSize caps how many applications one request can ask for. The panel
// asks for every application at once to offer them as choices.
const maxPageSize = 500

// maxURIs caps how many redirect URIs of each kind an application may have.
const maxURIs = 20

// applicationRequest is the body of the create and update endpoints. The
// names are RFC 7591's client metadata where it has one.
//
// Type is read on create only: it decides whether the app is a public client,
// which cannot change once copies of it are in use. A lifetime left at zero
// takes the default.
type applicationRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=255"`
	Type        string `json:"type"`

	LogoURI   string `json:"logo_uri" validate:"max=512"`
	ClientURI string `json:"client_uri" validate:"max=512"`
	PolicyURI string `json:"policy_uri" validate:"max=512"`
	TosURI    string `json:"tos_uri" validate:"max=512"`

	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	RedirectURIs            []string `json:"redirect_uris" validate:"dive,max=512"`
	PostLogoutRedirectURIs  []string `json:"post_logout_redirect_uris" validate:"dive,max=512"`
	Scopes                  []string `json:"scopes"`
	RequirePKCE             *bool    `json:"require_pkce"`

	AccessTokenLifetime  int `json:"access_token_lifetime"`
	IDTokenLifetime      int `json:"id_token_lifetime"`
	RefreshTokenLifetime int `json:"refresh_token_lifetime"`

	// The flags may be left out: a new application is then enabled, requires
	// PKCE and puts roles in its tokens; an existing one keeps its settings.
	AssertRoles           *bool `json:"assert_roles"`
	RequireRoleAssignment *bool `json:"require_role_assignment"`
	Enabled               *bool `json:"enabled"`
	// AllowRegistration may be left out too: a new application then offers
	// registration on its sign-in page.
	AllowRegistration *bool `json:"allow_registration"`

	// LoginFlowID is the flow this application signs people in with. An
	// empty string puts it back on the default flow; leaving it out leaves
	// the flow it has.
	LoginFlowID *string `json:"login_flow_id"`
}

// clean tidies what can be tidied, so the rules see the values that would
// actually be stored.
func (r *applicationRequest) clean() {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
	r.Type = strings.TrimSpace(r.Type)
	r.LogoURI = strings.TrimSpace(r.LogoURI)
	r.ClientURI = strings.TrimSpace(r.ClientURI)
	r.PolicyURI = strings.TrimSpace(r.PolicyURI)
	r.TosURI = strings.TrimSpace(r.TosURI)
	r.TokenEndpointAuthMethod = strings.TrimSpace(r.TokenEndpointAuthMethod)

	if r.AccessTokenLifetime == 0 {
		r.AccessTokenLifetime = model.DefaultAccessTokenLifetime
	}
	if r.IDTokenLifetime == 0 {
		r.IDTokenLifetime = model.DefaultIDTokenLifetime
	}
	if r.RefreshTokenLifetime == 0 {
		r.RefreshTokenLifetime = model.DefaultRefreshTokenLifetime
	}
}

// authorizeRequest is the body of the endpoint that authorises an application
// for an API: the ids of the API's scopes it may ask for.
type authorizeRequest struct {
	Scopes []uuid.UUID `json:"scopes"`
}

// previewRequest is a token request to evaluate: who for (nil for the
// application itself), which API, and the scope parameter as an application
// would send it.
type previewRequest struct {
	UserID   *uuid.UUID `json:"user_id"`
	Audience string     `json:"audience"`
	Scope    string     `json:"scope"`
}

// listQuery reads the search box, the filters and the page out of the query
// string, holding each to something sensible.
func listQuery(c *gin.Context) store.ApplicationQuery {
	q := store.ApplicationQuery{
		Search:  c.Query("search"),
		Limit:   query.Int(c, "limit", 50, maxPageSize),
		Offset:  query.Int(c, "offset", 0, query.MaxOffset),
		Enabled: query.Bool(c, "enabled"),
	}

	if kind := model.ApplicationType(c.Query("type")); kind.Valid() {
		q.Type = kind
	}

	return q
}
