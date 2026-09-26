package languages

import (
	"net/http"

	"loginer/internal/api/respond"
	"loginer/internal/api/validate"
	"loginer/internal/model"
)

// targetType is what a language is called in the activity log.
const targetType = "language"

// What these endpoints refuse, as the Languages page shows it.
var (
	codeTaken          = respond.Define(http.StatusConflict, "language_code_taken", respond.Admin)
	nothingToCopy      = respond.Define(http.StatusBadRequest, "language_copy_missing", respond.Admin)
	keepADefault       = respond.Define(http.StatusBadRequest, "language_default_required", respond.Admin)
	defaultStays       = respond.Define(http.StatusBadRequest, "language_default_protected", respond.Admin)
	baseStays          = respond.Define(http.StatusBadRequest, "language_base_protected", respond.Admin)
	noSuchApp          = respond.Define(http.StatusNotFound, "translation_app_not_found", respond.Admin)
	translationTooLong = respond.Define(http.StatusBadRequest, "translation_too_long", respond.Admin)
)

func init() {
	// The code is a language tag, as the model holds it to — said as a rule
	// here so a wrong one is refused with a sentence the panel can translate.
	validate.Register("languagetag", func(value string) bool {
		return (model.Language{Code: value, Name: "-", Native: "-"}).Validate() == nil
	})
}

// createRequest is what adding a language sends.
//
// CopyFrom names where its text starts: another language here, whose text is
// copied — Portuguese for Brazilian Portuguese — or a language the server
// ships with that this installation has not got, which brings its shipped
// text back. Empty is a language with nothing translated yet, whose pages
// show the base language's text until somebody writes its own.
type createRequest struct {
	Code      string `json:"code" validate:"required,max=16,languagetag"`
	Name      string `json:"name" validate:"required,max=64"`
	Native    string `json:"native" validate:"required,max=64"`
	Enabled   *bool  `json:"enabled"`
	IsDefault *bool  `json:"is_default"`
	CopyFrom  string `json:"copy_from"`
}

// languageRequest is what an update sends.
//
// Every field is a pointer because this is a PATCH: nil is "not sent", and
// the stored value stands. The code is not among them — it is in the path,
// and it is what every reader's saved choice names.
type languageRequest struct {
	Name      *string `json:"name" validate:"omitnil,min=1,max=64"`
	Native    *string `json:"native" validate:"omitnil,min=1,max=64"`
	Enabled   *bool   `json:"enabled"`
	IsDefault *bool   `json:"is_default"`
	Position  *int    `json:"position" validate:"omitnil,min=0"`
}

// translationRequest is one language's whole text for one app, replacing
// what was there: what the editor saves, and what an imported file is.
//
// `$name` and `$native` may come along — a file a translator wrote carries
// them — and are ignored here: the names are the language's own settings.
type translationRequest struct {
	Messages map[string]string `json:"messages" validate:"required"`
}
