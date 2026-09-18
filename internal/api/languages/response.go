package languages

import (
	"time"

	"github.com/google/uuid"

	"xermess/internal/model"
	"xermess/locales"
)

// languageResponse is one language as the panel sees it: its settings, and
// how much of each app it translates.
type languageResponse struct {
	Code string `json:"code"`
	// Name is what the language is called in English, Native what it calls
	// itself.
	Name   string `json:"name"`
	Native string `json:"native"`

	// Coverage is how much of each app is translated, as a percentage of the
	// keys the base language has, and Missing how many keys each is short.
	// The keys are the app names, "id" and "console".
	Coverage map[locales.App]int `json:"coverage"`
	Missing  map[locales.App]int `json:"missing"`

	Enabled   bool `json:"enabled"`
	IsDefault bool `json:"is_default"`
	Position  int  `json:"position"`

	// Base marks the language every other is a translation of: it cannot be
	// turned off or removed.
	Base bool `json:"base"`

	// Shipped says the server ships a translation of this language, so a key
	// a release adds reaches it on the next start.
	Shipped bool `json:"shipped"`

	UpdatedAt time.Time `json:"updated_at"`
}

// shippedResponse is a language the server ships with that this installation
// does not have — one somebody removed, or one that shipped after it started
// — which the panel offers to add back.
type shippedResponse struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Native string `json:"native"`
}

// listResponse is what the list endpoint answers with.
type listResponse struct {
	Languages []languageResponse `json:"languages"`
	// Apps are the two catalogs a language is counted against, so the panel
	// names the columns from the server rather than from a list of its own.
	Apps    []locales.App     `json:"apps"`
	Shipped []shippedResponse `json:"shipped"`
}

// response is what the single-language endpoints answer with.
type response struct {
	Language languageResponse `json:"language"`
}

// translationResponse is one language's text for one app, as the editor
// needs it: the keys there are, the base language's text beside each, and
// what this language has.
type translationResponse struct {
	App  locales.App `json:"app"`
	Keys []string    `json:"keys"`
	// Base is the text the base language shows, every key filled in: what
	// the translator is translating, and what a missing key falls back to.
	Base     map[string]string `json:"base"`
	Messages map[string]string `json:"messages"`
}

// savedResponse is what saving a translation answers with: the language as it
// now counts, and how many keys of what was sent were not ones the app looks
// up — a file from an older release, or a typo.
type savedResponse struct {
	Language languageResponse `json:"language"`
	Ignored  int              `json:"ignored"`
}

// panelLanguage is one language the panel itself can be shown in, and
// panelTextResponse its text.
type panelLanguage struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Native string `json:"native"`
}

type panelTextResponse struct {
	Language panelLanguage     `json:"language"`
	Messages map[string]string `json:"messages"`
}

func newLanguageResponse(language model.Language, text map[string]map[string]string, shipped map[string]bool) languageResponse {
	out := languageResponse{
		Code:      language.Code,
		Name:      language.Name,
		Native:    language.Native,
		Coverage:  map[locales.App]int{},
		Missing:   map[locales.App]int{},
		Enabled:   language.Enabled,
		IsDefault: language.IsDefault,
		Position:  language.Position,
		Base:      language.Code == model.BaseLanguage,
		Shipped:   shipped[language.Code],
		UpdatedAt: language.UpdatedAt,
	}

	for _, app := range locales.Apps {
		out.Coverage[app], out.Missing[app] = locales.Coverage(app, text[string(app)])
	}

	return out
}

func newListResponse(languages []model.Language, text map[uuid.UUID]map[string]map[string]string, files []locales.File) listResponse {
	shipped := make(map[string]bool, len(files))
	for _, file := range files {
		shipped[file.Code] = true
	}

	out := listResponse{
		Languages: make([]languageResponse, 0, len(languages)),
		Apps:      locales.Apps,
		Shipped:   []shippedResponse{},
	}

	have := make(map[string]bool, len(languages))
	for _, language := range languages {
		have[language.Code] = true
		out.Languages = append(out.Languages, newLanguageResponse(language, text[language.ID], shipped))
	}

	for _, file := range files {
		if !have[file.Code] {
			out.Shipped = append(out.Shipped, shippedResponse{Code: file.Code, Name: file.Name, Native: file.Native})
		}
	}

	return out
}
