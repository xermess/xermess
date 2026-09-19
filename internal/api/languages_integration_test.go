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

	for _, code := range []string{"en", "ru"} {
		language := list.find(code)
		switch {
		case language == nil:
			t.Fatalf("%s was not imported: %+v", code, list.Languages)
		case language.Coverage["id"] != 100:
			t.Errorf("%s covers %d%% of the sign-in pages, want all of them", code, language.Coverage["id"])
		case language.Enabled != (code == "en"):
			t.Errorf("%s enabled = %v, want only the base language offered", code, language.Enabled)
		}
	}

	if got := list.find("ru").Coverage["console"]; got != 100 {
		t.Errorf("ru covers %d%% of the panel, want all of it", got)
	}

	if len(list.Shipped) != 0 {
		t.Errorf("shipped but not installed = %v, want none on a fresh installation", list.Shipped)
	}

	if status := s.public("/languages/ru", nil); status != http.StatusNotFound {
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

	// The panel is shown in English and Russian only: a language an
	// installation adds is a sign-in language and nothing more.
	if _, counted := created.Language.Coverage["console"]; counted {
		t.Error("de was counted against the panel, which is not shown in it")
	}
	super.must(http.StatusNotFound, http.MethodGet, "/languages/de/translations/console", nil, nil)
	super.must(http.StatusNotFound, http.MethodPut, "/languages/de/translations/console", map[string]any{
		"messages": map[string]string{"nav.languages": "Sprachen"},
	}, nil)

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

	// The panel is drawn before anybody signs in, and only ever in English or
	// Russian — even with German offered to users.
	var panel struct {
		Languages []struct{ Code string } `json:"languages"`
	}
	s.client().must(http.StatusOK, http.MethodGet, "/panel/languages", nil, &panel)

	if len(panel.Languages) != 2 || panel.Languages[0].Code != "en" || panel.Languages[1].Code != "ru" {
		t.Errorf("the panel can be shown in %+v, want en and ru", panel.Languages)
	}
	s.client().must(http.StatusNotFound, http.MethodGet, "/panel/languages/de", nil, nil)

	// Read once more — from the cache, when there is one — then reword it: the
	// sign-in pages have the new text on the very next read, not an hour on.
	s.public("/languages/de", &text)
	super.must(http.StatusOK, http.MethodPut, "/languages/de/translations/id", map[string]any{
		"messages": map[string]string{"action.sign_in": "Einloggen"},
	}, nil)
	s.public("/languages/de", &text)
	if text.Messages["action.sign_in"] != "Einloggen" {
		t.Errorf("after rewording, action.sign_in = %q, want the new text", text.Messages["action.sign_in"])
	}

	// English is what every other language falls back to, so rewording it
	// reaches German's untranslated keys at once too.
	super.must(http.StatusOK, http.MethodPut, "/languages/en/translations/id", map[string]any{
		"messages": map[string]string{"login.title": "Welcome back"},
	}, nil)
	s.public("/languages/de", &text)
	if text.Messages["login.title"] != "Welcome back" {
		t.Errorf("German's fallback for login.title = %q, want the reworded English", text.Messages["login.title"])
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

	// A shipped language removed is offered back, and comes back whole — for
	// the sign-in pages and the panel both.
	super.must(http.StatusNoContent, http.MethodDelete, "/languages/ru", nil, nil)
	super.must(http.StatusOK, http.MethodGet, "/languages", nil, &list)

	if len(list.Shipped) != 1 || list.Shipped[0].Code != "ru" {
		t.Errorf("shipped but not installed = %+v, want ru", list.Shipped)
	}

	super.must(http.StatusCreated, http.MethodPost, "/languages", map[string]any{
		"code": "ru", "copy_from": "ru",
	}, &created)

	if created.Language.Native == "" || created.Language.Coverage["id"] != 100 || created.Language.Coverage["console"] != 100 {
		t.Errorf("ru brought back = %+v, want its names and all its text", created.Language)
	}

	var panelText struct {
		Messages map[string]string `json:"messages"`
	}
	s.client().must(http.StatusOK, http.MethodGet, "/panel/languages/ru", nil, &panelText)
	if panelText.Messages["nav.languages"] == "" || panelText.Messages["nav.languages"] == "Languages" {
		t.Errorf("nav.languages in Russian = %q", panelText.Messages["nav.languages"])
	}
}
