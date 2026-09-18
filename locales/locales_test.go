package locales

import "testing"

// Every shipped language has to name itself, and the base language has to be
// there at all: it is what a missing translation falls back to, and what
// coverage is a percentage of.
func TestShipped(t *testing.T) {
	shipped, err := Shipped()
	if err != nil {
		t.Fatalf("Shipped() = %v", err)
	}

	if len(shipped) == 0 || shipped[0].Code != Base {
		t.Fatalf("the shipped languages start with %v, want the base language first", shipped)
	}

	for _, language := range shipped {
		t.Run(language.Code, func(t *testing.T) {
			switch {
			case language.Name == "" || language.Name == language.Code:
				t.Errorf("$name is missing: a language has to say what it is called")
			case language.Native == "":
				t.Error("$native is missing")
			}

			for app, messages := range language.Messages {
				if _, meta := messages["$name"]; meta {
					t.Errorf("%s still carries $name among its messages for %s", language.Code, app)
				}
			}
		})
	}
}

// A language that ships has to be complete: the first start imports it, and
// an installation that offers it should not be offering half a page.
func TestTranslationsAreComplete(t *testing.T) {
	shipped, err := Shipped()
	if err != nil {
		t.Fatal(err)
	}

	for _, language := range shipped {
		for _, app := range Apps {
			if percent, missing := Coverage(app, language.Messages[app]); missing > 0 {
				t.Errorf("%s is %d%% of %s, short %d keys; add them to its file",
					language.Code, percent, app, missing)
			}
		}
	}
}

func TestCoverage(t *testing.T) {
	keys := Keys(ID)
	if len(keys) < 2 {
		t.Fatal("the base language has too few keys to test coverage with")
	}

	tests := []struct {
		name        string
		messages    Messages
		wantMissing int
	}{
		{name: "nothing translated", messages: Messages{}, wantMissing: len(keys)},
		{name: "one key", messages: Messages{keys[0]: "x"}, wantMissing: len(keys) - 1},
		{name: "a key that is only spaces", messages: Messages{keys[0]: "  "}, wantMissing: len(keys)},
		{name: "a key nothing looks up", messages: Messages{"nothing.here": "x"}, wantMissing: len(keys)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, missing := Coverage(ID, tt.messages)
			if missing != tt.wantMissing {
				t.Errorf("missing = %d, want %d", missing, tt.wantMissing)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	const key = "action.sign_in"

	shippedEnglish := Resolve(ID)[key]
	if shippedEnglish == "" {
		t.Fatalf("the base language has no %s", key)
	}

	tests := []struct {
		name   string
		layers []Messages
		want   string
	}{
		{name: "the language has the key", layers: []Messages{{key: "Кирүү"}, {key: "Log in"}}, want: "Кирүү"},
		{name: "the language is missing it", layers: []Messages{{}, {key: "Log in"}}, want: "Log in"},
		{name: "an empty translation is missing", layers: []Messages{{key: " "}, {key: "Log in"}}, want: "Log in"},
		{name: "nobody has it", layers: []Messages{{}, {}}, want: shippedEnglish},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Resolve(ID, tt.layers...)[key]; got != tt.want {
				t.Errorf("Resolve()[%q] = %q, want %q", key, got, tt.want)
			}
		})
	}

	if _, kept := Resolve(ID, Messages{"nothing.here": "x"})["nothing.here"]; kept {
		t.Error("Resolve kept a key the base language does not have")
	}

	if got, want := len(Resolve(Console)), len(Keys(Console)); got != want {
		t.Errorf("Resolve(Console) has %d keys, want every one of the %d", got, want)
	}
}

func TestKnown(t *testing.T) {
	key := Keys(ID)[0]

	kept, dropped := Known(ID, Messages{
		"$name":        "Deutsch",
		key:            "Hallo",
		Keys(ID)[1]:    "  ",
		"nothing.here": "x",
	})

	if len(kept) != 1 || kept[key] != "Hallo" {
		t.Errorf("kept = %v, want only %s", kept, key)
	}
	if dropped != 1 {
		t.Errorf("dropped = %d, want 1: the unknown key, not the name or the empty one", dropped)
	}
}

func TestParseApp(t *testing.T) {
	for _, app := range Apps {
		if got, ok := ParseApp(string(app)); !ok || got != app {
			t.Errorf("ParseApp(%q) = %q, %v", app, got, ok)
		}
	}

	if _, ok := ParseApp("../id"); ok {
		t.Error("ParseApp accepted something that is not an app")
	}
}
