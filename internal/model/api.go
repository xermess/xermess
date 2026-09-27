package model

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// API is a resource server: a service applications ask for access tokens to
// call, Auth0's API and Keycloak's audience.
//
// It is how tokens are down-scoped. A token is issued for one API at a time —
// its audience — and carries only the scopes of that API the application is
// allowed to ask for and, when the API enforces roles, the user's roles grant.
// An API that trusts a token with its own audience therefore never sees a
// scope meant for another service.
type API struct {
	Base

	Name string `gorm:"size:100;not null" json:"name"`
	// Identifier is the audience: the value of a token's aud claim, and what
	// an application names to ask for a token for this API. It never
	// changes, since every service checking tokens compares against it.
	Identifier  string `gorm:"size:255;not null;uniqueIndex" json:"identifier"`
	Description string `gorm:"size:255" json:"description"`

	// EnforceRoles grants a user's token only the scopes their roles grant.
	// Off, the application's allowed scopes decide, for an API that makes its
	// own decisions about users.
	EnforceRoles bool `gorm:"not null" json:"enforce_roles"`

	// SigningAlgorithm is how access tokens for the API are signed. It is
	// asymmetric, so the API checks tokens with the public keys this server
	// publishes and never holds a key that could make one.
	SigningAlgorithm string `gorm:"type:varchar(16);not null" json:"signing_algorithm"`

	// TokenLifetime is how long an access token for the API lasts, in
	// seconds. Zero leaves it to the application's access token lifetime.
	TokenLifetime int `gorm:"not null" json:"token_lifetime"`

	// AllowOfflineAccess lets an application get a refresh token alongside
	// an access token for the API.
	AllowOfflineAccess bool `gorm:"not null" json:"allow_offline_access"`

	Scopes []APIScope `gorm:"constraint:OnDelete:CASCADE" json:"scopes,omitempty"`
}

// TableName pins the table name.
func (API) TableName() string {
	return "apis"
}

// APIScope is one thing a token for an API may allow — "orders:read" — which
// the API checks for.
type APIScope struct {
	Base

	APIID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_api_scopes_api_name,priority:1" json:"api_id"`
	Name        string    `gorm:"size:128;not null;uniqueIndex:idx_api_scopes_api_name,priority:2" json:"name"`
	Description string    `gorm:"size:255" json:"description"`

	// IsDefault adds the scope to every access token for the API, whether or
	// not it was asked for — still only when the application is allowed it
	// and, if the API enforces roles, the user's roles grant it.
	IsDefault bool `gorm:"not null" json:"is_default"`
}

// The algorithms access tokens can be signed with. All are asymmetric: HS256
// is left out on purpose, since every API checking tokens would have to hold
// the secret that makes them.
const (
	AlgRS256 = "RS256"
	AlgPS256 = "PS256"
	AlgES256 = "ES256"
)

// SigningAlgorithms lists every algorithm, the default first.
var SigningAlgorithms = []string{AlgRS256, AlgPS256, AlgES256}

// Bounds on an API's own token lifetime, when it sets one.
const (
	minAPITokenLifetime = 60
	maxAPITokenLifetime = 24 * 60 * 60
)

// TableName pins the table name.
func (APIScope) TableName() string {
	return "api_scopes"
}

// ApplicationAPI says an application may ask for tokens for an API: Auth0's
// authorised application. Without it, a request naming the API as audience is
// refused outright.
type ApplicationAPI struct {
	ApplicationID uuid.UUID    `gorm:"type:uuid;primaryKey" json:"application_id"`
	Application   *Application `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	APIID         uuid.UUID    `gorm:"type:uuid;primaryKey" json:"api_id"`
	API           *API         `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt     time.Time    `json:"created_at"`
}

// TableName pins the table name.
func (ApplicationAPI) TableName() string {
	return "application_apis"
}

// ApplicationAPIScope is one scope of an API an application may ask for: the
// ceiling on what any token for the application carries, whoever the user.
type ApplicationAPIScope struct {
	ApplicationID uuid.UUID    `gorm:"type:uuid;primaryKey" json:"application_id"`
	Application   *Application `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	APIScopeID    uuid.UUID    `gorm:"type:uuid;primaryKey" json:"api_scope_id"`
	APIScope      *APIScope    `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// TableName pins the table name.
func (ApplicationAPIScope) TableName() string {
	return "application_api_scopes"
}

// scopeNamePattern is what an API scope may be called: lower case words
// joined by colons, dots, dashes or underscores — "orders:read",
// "read:orders", "billing.invoices.export".
var scopeNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*([:._-][a-z0-9]+)*$`)

// Validate reports the first thing wrong with an API and its scopes.
func (a API) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("name is required")
	}

	if a.Identifier == "" {
		return fmt.Errorf("identifier is required")
	}
	if strings.ContainsAny(a.Identifier, " \t\r\n") {
		return fmt.Errorf("identifier must not contain spaces")
	}
	if len(a.Identifier) > 255 {
		return fmt.Errorf("identifier must be at most 255 characters")
	}

	if !slices.Contains(SigningAlgorithms, a.SigningAlgorithm) {
		return fmt.Errorf("signing_algorithm must be one of: %s", strings.Join(SigningAlgorithms, ", "))
	}

	if a.TokenLifetime != 0 && (a.TokenLifetime < minAPITokenLifetime || a.TokenLifetime > maxAPITokenLifetime) {
		return fmt.Errorf("token_lifetime must be between %d and %d seconds, or 0 to use the application's", minAPITokenLifetime, maxAPITokenLifetime)
	}

	seen := map[string]bool{}
	for _, scope := range a.Scopes {
		switch {
		case !scopeNamePattern.MatchString(scope.Name) || len(scope.Name) > 128:
			return fmt.Errorf("scopes: %q must be lower case words joined by : . - or _, such as orders:read", scope.Name)
		case slices.Contains(Scopes, scope.Name):
			return fmt.Errorf("scopes: %q is an OpenID Connect scope and cannot be an API scope", scope.Name)
		case seen[scope.Name]:
			return fmt.Errorf("scopes: %q is listed twice", scope.Name)
		case len(scope.Description) > 255:
			return fmt.Errorf("scopes: the description of %q must be at most 255 characters", scope.Name)
		}
		seen[scope.Name] = true
	}

	return nil
}
