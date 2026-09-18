package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

// A language, end to end: imported from the shipped files on the first start,
// added and translated in the panel, offered to the sign-in pages and served
// to them with its gaps filled in, and removed again.

type languageBody struct {
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	Native    string         `json:"native"`
	Coverage  map[string]int `json:"coverage"`
	Missing   map[string]int `json:"missing"`
	Enabled   bool           `json:"enabled"`
	IsDefault bool           `json:"is_default"`
}

type languageList struct {
	Languages []languageBody `json:"languages"`
	Shipped   []struct {
		Code string `json:"code"`
	} `json:"shipped"`
}

func (l languageList) find(code string) *languageBody {
	for i := range l.Languages {
		if l.Languages[i].Code == code {
			return &l.Languages[i]
		}
	}

	return nil
}

// public asks the public server, as the sign-in pages do.
func (s *liveServer) public(path string, out any) int {
	s.t.Helper()

	res, err := http.Get(s.root + "/api/v1/account" + path)
	if err != nil {
		s.t.Fatal(err)
	}
	defer res.Body.Close()

	if out != nil && res.StatusCode == http.StatusOK {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			s.t.Fatal(err)
		}
	}

	return res.StatusCode
}

func TestLiveLanguages(t *testing.T) {
	s := newLiveServer(t)
	super := s.superAdmin()

	// The first start imported every shipped language, whole, and offers only
	// the base one.
	var list languageList
	super.must(http.StatusOK, http.MethodGet, "/languages", nil, &list)

	for _, code := range []string{"en", "ky", "ru"} {
		language := list.find(code)
		switch {
		case language == nil:
			t.Fatalf("%s was not imported: %+v", code, list.Languages)
		case language.Coverage["id"] != 100 || language.Coverage["console"] != 100:
			t.Errorf("%s coverage = %v, want all of both apps", code, language.Coverage)
		case language.Enabled != (code == "en"):
			t.Errorf("%s enabled = %v, want only the base language offered", code, language.Enabled)
		}
	}

	if len(list.Shipped) != 0 {
		t.Errorf("shipped but not installed = %v, want none on a fresh installation", list.Shipped)
	}

	if status := s.public("/languages/ky", nil); status != http.StatusNotFound {
		t.Errorf("a language that is off = %d to the sign-in pages, want 404", status)
	}

	// A language that did not ship, added from nothing and translated.
	var created struct{ Language languageBody }
	super.must(http.StatusCreated, http.MethodPost, "/languages", map[string]any{
		"code": "de", "name": "German", "native": "Deutsch",
	}, &created)

	if created.Language.Coverage["id"] != 0 {
		t.Errorf("a new language covers %d%% of the sign-in pages, want nothing yet", created.Language.Coverage["id"])
	}

	super.must(http.StatusConflict, http.MethodPost, "/languages", map[string]any{
		"code": "DE", "name": "German", "native": "Deutsch",
	}, nil)

	var saved struct {
		Language languageBody `json:"language"`
		Ignored  int          `json:"ignored"`
	}
	super.must(http.StatusOK, http.MethodPut, "/languages/de/translations/id", map[string]any{
		"messages": map[string]string{"action.sign_in": "Anmelden", "nothing.here": "x", "$name": "Deutsch"},
	}, &saved)

	if saved.Ignored != 1 {
		t.Errorf("ignored = %d, want the one key nothing looks up", saved.Ignored)
	}
	if saved.Language.Missing["id"] != created.Language.Missing["id"]-1 {
		t.Errorf("missing = %d after translating one of %d, want one fewer",
			saved.Language.Missing["id"], created.Language.Missing["id"])
	}

	super.must(http.StatusNotFound, http.MethodPut, "/languages/de/translations/mobile", map[string]any{
		"messages": map[string]string{},
	}, nil)

	// Offered, the sign-in pages list it and get every key: its own where it
	// has one, English where it does not.
	super.must(http.StatusOK, http.MethodPatch, "/languages/de", map[string]any{"enabled": true}, nil)

	var offered struct {
		Languages []struct{ Code string } `json:"languages"`
		Default   string                  `json:"default"`
	}
	s.public("/languages", &offered)

	if len(offered.Languages) != 2 || offered.Languages[1].Code != "de" || offered.Default != "en" {
		t.Errorf("offered = %+v, want en then de, en the default", offered)
	}

	var text struct {
		Messages map[string]string `json:"messages"`
	}
	if status := s.public("/languages/de", &text); status != http.StatusOK {
		t.Fatalf("the text of an offered language = %d", status)
	}
	if text.Messages["action.sign_in"] != "Anmelden" {
		t.Errorf("action.sign_in = %q, want the German", text.Messages["action.sign_in"])
	}
	if text.Messages["login.title"] == "" {
		t.Error("a key German does not have came back empty rather than in English")
	}

	// The two languages nothing may remove: the base, and the default.
	super.must(http.StatusBadRequest, http.MethodDelete, "/languages/en", nil, nil)
	super.must(http.StatusOK, http.MethodPatch, "/languages/de", map[string]any{"is_default": true}, nil)
	super.must(http.StatusBadRequest, http.MethodDelete, "/languages/de", nil, nil)
	super.must(http.StatusBadRequest, http.MethodPatch, "/languages/de", map[string]any{"enabled": false, "is_default": false}, nil)

	super.must(http.StatusOK, http.MethodGet, "/languages", nil, &list)
	if list.find("en").IsDefault || !list.find("de").IsDefault {
		t.Errorf("making de the default left en = %+v, de = %+v", list.find("en"), list.find("de"))
	}

	super.must(http.StatusOK, http.MethodPatch, "/languages/en", map[string]any{"is_default": true}, nil)
	super.must(http.StatusNoContent, http.MethodDelete, "/languages/de", nil, nil)

	// A shipped language removed is offered back, and comes back whole.
	super.must(http.StatusNoContent, http.MethodDelete, "/languages/ky", nil, nil)
	super.must(http.StatusOK, http.MethodGet, "/languages", nil, &list)

	if len(list.Shipped) != 1 || list.Shipped[0].Code != "ky" {
		t.Errorf("shipped but not installed = %+v, want ky", list.Shipped)
	}

	super.must(http.StatusCreated, http.MethodPost, "/languages", map[string]any{
		"code": "ky", "copy_from": "ky",
	}, &created)

	if created.Language.Native == "" || created.Language.Coverage["console"] != 100 {
		t.Errorf("ky brought back = %+v, want its names and all its text", created.Language)
	}

	// The panel is drawn before anybody signs in, in any language with some
	// of the panel translated.
	var panel struct {
		Languages []struct{ Code string } `json:"languages"`
	}
	s.client().must(http.StatusOK, http.MethodGet, "/panel/languages", nil, &panel)

	if len(panel.Languages) != 3 {
		t.Errorf("the panel can be shown in %+v, want en, ky and ru", panel.Languages)
	}

	var panelText struct {
		Messages map[string]string `json:"messages"`
	}
	s.client().must(http.StatusOK, http.MethodGet, "/panel/languages/ru", nil, &panelText)
	if panelText.Messages["nav.languages"] == "" || panelText.Messages["nav.languages"] == "Languages" {
		t.Errorf("nav.languages in Russian = %q", panelText.Messages["nav.languages"])
	}
}
