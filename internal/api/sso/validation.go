package sso

import (
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"

	"xermess/internal/api/respond"
	"xermess/internal/api/validate"
	"xermess/internal/jose"
	"xermess/internal/model"
	"xermess/internal/oidc"
)

// applyTo checks the request and copies it onto a connection.
//
// `creating` says whether the slug and the protocol are read: neither may
// change afterwards, since both are in the addresses the identity provider
// was given. Everything else is copied only where the request mentioned it,
// so a change to one setting is a change to one setting.
func (r *connectionRequest) applyTo(connection *model.SSOConnection, sealer *jose.Sealer, creating bool) error {
	if err := validate.Struct(r); err != nil {
		return err
	}

	if creating {
		connection.Protocol = model.SSOProtocol(validate.Text(r.Protocol, ""))
		connection.Slug = slugFrom(validate.Text(r.Slug, ""), validate.Text(r.Name, ""))

		if connection.Protocol == "" {
			return required("protocol")
		}
		if !slugPattern.MatchString(connection.Slug) {
			return invalid("slug")
		}
	}

	connection.Name = validate.Text(r.Name, connection.Name)
	connection.Enabled = validate.Flag(r.Enabled, connection.Enabled)
	connection.EnforceDomains = validate.Flag(r.EnforceDomains, connection.EnforceDomains)
	connection.ShowOnLogin = validate.Flag(r.ShowOnLogin, connection.ShowOnLogin)

	if r.Domains != nil {
		domains := model.StringList{}
		for _, domain := range *r.Domains {
			domain = model.NormalizeDomain(domain)
			if domain == "" || slices.Contains(domains, domain) {
				continue
			}
			if !model.ValidDomain(domain) {
				return domainInvalid.With("domain", domain)
			}
			domains = append(domains, domain)
		}
		connection.Domains = domains
	}

	connection.Issuer = strings.TrimRight(validate.Text(r.Issuer, connection.Issuer), "/")
	connection.ClientID = validate.Text(r.ClientID, connection.ClientID)

	// A secret that was sent replaces the stored one; none leaves it be.
	if r.ClientSecret != nil && strings.TrimSpace(*r.ClientSecret) != "" {
		sealed, err := sealer.SealBytes([]byte(strings.TrimSpace(*r.ClientSecret)))
		if err != nil {
			return err
		}
		connection.ClientSecret = sealed
	}

	if r.Scopes != nil {
		scopes := model.StringList{}
		for _, scope := range *r.Scopes {
			for _, piece := range strings.Fields(scope) {
				if !slices.Contains(scopes, piece) {
					scopes = append(scopes, piece)
				}
			}
		}
		connection.Scopes = scopes
	}

	connection.MetadataURL = validate.Text(r.MetadataURL, connection.MetadataURL)
	if r.Metadata != nil {
		connection.Metadata = strings.TrimSpace(*r.Metadata)
	}
	connection.NameIDFormat = model.SSONameIDFormat(validate.Text(r.NameIDFormat, string(connection.NameIDFormat)))
	connection.SignRequests = validate.Flag(r.SignRequests, connection.SignRequests)

	connection.Matching = model.SSOMatching(validate.Text(r.Matching, string(connection.Matching)))
	connection.CreateUsers = validate.Flag(r.CreateUsers, connection.CreateUsers)
	connection.SyncProfile = validate.Flag(r.SyncProfile, connection.SyncProfile)
	connection.SyncRoles = validate.Flag(r.SyncRoles, connection.SyncRoles)

	connection.EmailAttribute = validate.Text(r.EmailAttribute, connection.EmailAttribute)
	connection.FirstNameAttribute = validate.Text(r.FirstNameAttribute, connection.FirstNameAttribute)
	connection.LastNameAttribute = validate.Text(r.LastNameAttribute, connection.LastNameAttribute)
	connection.GroupsAttribute = validate.Text(r.GroupsAttribute, connection.GroupsAttribute)

	if r.RoleMappings != nil {
		mappings := model.SSORoleMappings{}
		for _, mapping := range *r.RoleMappings {
			id, err := uuid.Parse(mapping.RoleID)
			group := strings.TrimSpace(mapping.Group)
			if err != nil || group == "" || len(group) > 255 {
				return mappingInvalid.With("group", group)
			}
			mappings = append(mappings, model.SSORoleMapping{Group: group, RoleID: id})
		}
		connection.RoleMappings = mappings
	}

	return settle(connection)
}

// settle checks what a connection has to have, whoever wrote it, with a
// problem the panel can translate for each; the model's own check is the
// last word after them.
func settle(connection *model.SSOConnection) error {
	// Without domains no address leads to the connection, so it needs its
	// button — and there is nobody to require it of.
	if len(connection.Domains) == 0 && (connection.EnforceDomains || !connection.ShowOnLogin) {
		return unreachable.Fault(nil)
	}

	switch connection.Protocol {
	case model.SSOProtocolOIDC:
		parsed, err := url.Parse(connection.Issuer)
		switch {
		case connection.Issuer == "":
			return required("issuer")
		case err != nil || parsed.Host == "" || parsed.Scheme != "https" && !localhost(parsed):
			return issuerInvalid.Fault(nil)
		case connection.ClientID == "":
			return required("client_id")
		}
	case model.SSOProtocolSAML:
		if connection.Metadata == "" && connection.MetadataURL == "" {
			return required("metadata")
		}
		if connection.NameIDFormat == "" {
			connection.NameIDFormat = model.SSONameIDEmail
		}
	}

	if connection.Matching == "" {
		connection.Matching = model.SSOMatchLink
	}
	if connection.Domains == nil {
		connection.Domains = model.StringList{}
	}
	if connection.RoleMappings == nil {
		connection.RoleMappings = model.SSORoleMappings{}
	}

	if err := connection.Validate(); err != nil {
		return respond.Fault{Status: 400, Message: err.Error()}
	}

	return nil
}

// localhost is an identity provider on this machine, which may be plain
// http: that is how one is tried out.
func localhost(u *url.URL) bool {
	host := u.Hostname()
	return u.Scheme == "http" && (host == "localhost" || host == "127.0.0.1" || host == "::1")
}

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?$`)
	notSlug     = regexp.MustCompile(`[^a-z0-9]+`)

	requiredField = func() respond.Problem { p, _ := respond.Lookup("validation.required"); return p }()
	invalidField  = func() respond.Problem { p, _ := respond.Lookup("validation.invalid"); return p }()
)

// slugFrom is the slug sent, or one made from the name: "Acme Corp" is
// "acme-corp".
func slugFrom(sent, name string) string {
	slug := strings.ToLower(strings.TrimSpace(sent))
	if slug == "" {
		slug = strings.Trim(notSlug.ReplaceAllString(strings.ToLower(name), "-"), "-")
	}

	if len(slug) > 64 {
		slug = strings.TrimRight(slug[:64], "-")
	}

	return slug
}

func required(field string) respond.Fault { return requiredField.With("field", field) }
func invalid(field string) respond.Fault  { return invalidField.With("field", field) }

// unsupportedScopes are the scopes a connection would ask for that its
// provider does not list. A provider that lists none says nothing either way.
func unsupportedScopes(scopes, supported []string) []string {
	if len(supported) == 0 {
		return nil
	}

	var out []string
	for _, scope := range (model.SSOConnection{Scopes: scopes}).AskedScopes() {
		if !slices.Contains(supported, scope) {
			out = append(out, scope)
		}
	}

	return out
}

// providerOf is who a connection believes: an OpenID Connect issuer, or a
// SAML provider's entity ID. The subjects of its identities are only unique
// within it, so a connection pointed at another one starts them afresh.
func providerOf(connection *model.SSOConnection) string {
	if connection.Protocol == model.SSOProtocolOIDC {
		return connection.Issuer
	}

	descriptor, err := oidc.ParseSAMLMetadata([]byte(connection.Metadata))
	if err != nil {
		return ""
	}

	return descriptor.EntityID
}
