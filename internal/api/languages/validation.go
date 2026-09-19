package languages

import (
	"net/http"
	"strings"

	"xermess/internal/api/respond"
	"xermess/internal/api/validate"
	"xermess/internal/model"
)

// applyTo checks the request and copies it onto a language.
//
// Marking a language as the default is also turning it on, since the default
// is what somebody sees before choosing and has to be among the choices.
func (r *languageRequest) applyTo(language *model.Language) error {
	if err := validate.Struct(r); err != nil {
		return err
	}

	if r.Name != nil {
		language.Name = strings.TrimSpace(*r.Name)
	}
	if r.Native != nil {
		language.Native = strings.TrimSpace(*r.Native)
	}
	if r.Position != nil {
		language.Position = *r.Position
	}

	language.Enabled = validate.Flag(r.Enabled, language.Enabled)
	language.IsDefault = validate.Flag(r.IsDefault, language.IsDefault)

	return settle(language)
}

// newLanguage checks a create request and makes the language it describes.
func (r *createRequest) newLanguage() (*model.Language, error) {
	r.Code = normalizeCode(r.Code)
	r.Name = strings.TrimSpace(r.Name)
	r.Native = strings.TrimSpace(r.Native)

	if err := validate.Struct(r); err != nil {
		return nil, err
	}

	language := &model.Language{
		Code:      normalizeCode(r.Code),
		Name:      strings.TrimSpace(r.Name),
		Native:    strings.TrimSpace(r.Native),
		Enabled:   validate.Flag(r.Enabled, false),
		IsDefault: validate.Flag(r.IsDefault, false),
	}

	return language, settle(language)
}

// settle applies the rules a language is always held to, whoever wrote it,
// and validates what is left.
func settle(language *model.Language) error {
	if language.IsDefault {
		language.Enabled = true
	}

	// The base language is always offered: a page drawn in it is the last
	// thing that still works when every other language is short of a key.
	if language.Code == model.BaseLanguage {
		language.Enabled = true
	}

	if err := language.Validate(); err != nil {
		return badRequest(err.Error())
	}

	return nil
}

// normalizeCode writes a language tag the way the standard does, so "PT-br"
// and "pt-BR" are one language rather than two: the language lower case, a
// four-letter script title case, and a region upper case.
func normalizeCode(code string) string {
	parts := strings.Split(strings.TrimSpace(code), "-")

	for i, part := range parts {
		switch {
		case i == 0:
			parts[i] = strings.ToLower(part)
		case len(part) == 4:
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		case len(part) == 2 || len(part) == 3 && part[0] >= '0' && part[0] <= '9':
			parts[i] = strings.ToUpper(part)
		default:
			parts[i] = strings.ToLower(part)
		}
	}

	return strings.Join(parts, "-")
}

// badRequest is a 400 carrying what was wrong with the request.
func badRequest(message string) error {
	return respond.Fault{Status: http.StatusBadRequest, Message: message}
}
