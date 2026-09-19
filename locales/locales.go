// Package locales is the translations this server ships with, and the one
// list of which keys there are.
//
// A shipped language is a JSON file: `locales/id/ru.json` is Russian for the
// sign-in pages, `locales/console/ru.json` is Russian for the admin panel —
// which only PanelLanguages have. They are embedded, so a deployment is one
// binary and carries them.
//
// A file is nested by screen — `{"login": {"title": "Sign in"}}` — so a
// translator reads one page's text together, and Flatten reads it into the
// dotted keys everything else speaks of: "login.title". The `error`
// namespace is the server's to fill (internal/api/respond): every problem it
// can answer with has its sentence there, and its English comes from here.
//
// What an installation actually serves lives in the database, not here. On
// the first start every file is imported (store.EnsureLanguages), and from
// then on the languages are the panel's: an administrator adds one, edits its
// text, imports a file somebody translated, or removes it, and none of that
// needs a build. The files keep two jobs after that:
//
//   - The base language's files are the contract. The keys `en.json` has are
//     the keys the apps look up, so coverage is counted against them, and its
//     text is what a key nobody has translated falls back to — even when the
//     database copy of English has lost it.
//   - A release that adds a key adds its translation to the shipped files, and
//     the next start copies the new key into the database for every language
//     that has a file, without touching what an administrator has written.
package locales

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
)

//go:embed id/*.json console/*.json
var files embed.FS

// App is one of the two apps a translation is for.
type App string

const (
	// ID is the sign-in pages and a user's own account, web/id.
	ID App = "id"
	// Console is the admin panel, web/console.
	Console App = "console"
)

// Apps is both of them, in the order the panel lists them.
var Apps = []App{ID, Console}

// ParseApp reads an app's name, as it appears in a path.
func ParseApp(name string) (App, bool) {
	for _, app := range Apps {
		if string(app) == name {
			return app, true
		}
	}

	return "", false
}

// PanelLanguages are the only languages the admin panel is shown in.
//
// The sign-in pages are for everybody who has an account, so any language
// an installation adds is theirs to offer. The panel is for the people who
// run the installation, and it is kept to these: every screen of it has to
// be translated and kept translated as it changes, which is work worth doing
// for two languages rather than for every one somebody adds. A language not
// named here has sign-in text and no panel text — the panel's text for it is
// not imported, not edited, not served, and not counted.
var PanelLanguages = []string{"en", "ru"}

// AppsFor is the apps a language is translated for: the sign-in pages for
// every language, and the panel for PanelLanguages.
func AppsFor(code string) []App {
	if slices.Contains(PanelLanguages, code) {
		return Apps
	}

	return []App{ID}
}

// ServesApp reports whether a language is translated for an app.
func ServesApp(code string, app App) bool {
	return slices.Contains(AppsFor(code), app)
}

// Base is the language every other is a translation of: the keys it has are
// the keys there are, and the text it holds is what a missing translation
// falls back to.
const Base = "en"

// Messages is one language's text for one app, by key.
type Messages = map[string]string

// File is one language this server ships with.
type File struct {
	// Code is the file's name: an IETF language tag, "ky" or "pt-BR".
	Code string

	// Name is what the language is called in English, and Native what it is
	// called in itself, both read from the file so a translator names their
	// own language.
	Name   string
	Native string

	// Messages is the text for each app, without the `$` keys that describe
	// the file. An app the language has no file for is left out.
	Messages map[App]Messages
}

// metaPrefix marks the keys that are about the file rather than in the
// interface: they are not counted, and never looked up.
const metaPrefix = "$"

// Shipped is every language there are files for, the base language first and
// the rest by name.
func Shipped() ([]File, error) {
	shipped, err := load()
	if err != nil {
		return nil, err
	}

	out := make([]File, len(shipped.files))
	copy(out, shipped.files)

	return out, nil
}

// Keys are the keys an app looks up, sorted: the base language's, without its
// metadata. Sorted rather than in file order, which also puts every key of
// one screen together — they share a prefix.
func Keys(app App) []string {
	shipped, err := load()
	if err != nil {
		return nil
	}

	return shipped.keys[app]
}

// Coverage is how much of an app a language's messages translate, as a
// percentage of the keys there are, and how many keys it is short. A key
// holding only spaces is not a translation.
func Coverage(app App, messages Messages) (percent, missing int) {
	keys := Keys(app)
	if len(keys) == 0 {
		return 100, 0
	}

	translated := 0
	for _, key := range keys {
		if strings.TrimSpace(messages[key]) != "" {
			translated++
		}
	}

	return translated * 100 / len(keys), len(keys) - translated
}

// Resolve is what an app is sent for one language: every key there is, each
// taken from the first of `layers` that has it, and from the shipped base
// language when none does.
//
// The layers are the language itself and then the base language as the
// database holds it, so an administrator's edit to the English text reaches
// every language still missing that key. The shipped file under them is what
// keeps a page from ever showing a bare key, whatever has been deleted.
//
// Keys that are not in the base language are dropped: nothing looks them up,
// and a file from an older release may still carry some.
func Resolve(app App, layers ...Messages) Messages {
	shipped, err := load()
	if err != nil {
		shipped = &catalog{}
	}

	base := shipped.base[app]
	out := make(Messages, len(shipped.keys[app]))

	for _, key := range shipped.keys[app] {
		out[key] = base[key]

		for _, layer := range layers {
			if value := layer[key]; strings.TrimSpace(value) != "" {
				out[key] = value
				break
			}
		}
	}

	return out
}

// Text is the base language's shipped text for one key of an app, and
// whether there is any. It is what the server says in English when it has to
// say something itself — the sentence beside an error's code.
func Text(app App, key string) (string, bool) {
	shipped, err := load()
	if err != nil {
		return "", false
	}

	text, ok := shipped.base[app][key]

	return text, ok
}

// Fill replaces each `{name}` in a text with the parameter of that name, the
// same way the apps do. A parameter nobody passed is left as it was written.
func Fill(text string, params map[string]any) string {
	if len(params) == 0 {
		return text
	}

	return placeholder.ReplaceAllStringFunc(text, func(whole string) string {
		value, ok := params[whole[1:len(whole)-1]]
		if !ok {
			return whole
		}

		return fmt.Sprint(value)
	})
}

// placeholder is a `{name}` in a text.
var placeholder = regexp.MustCompile(`\{(\w+)\}`)

// Known keeps the messages whose keys an app looks up and drops the rest,
// with how many it dropped. An empty value is dropped too: it is the same as
// no translation, and storing it would only make the file bigger.
func Known(app App, messages Messages) (Messages, int) {
	keys := make(map[string]bool, len(Keys(app)))
	for _, key := range Keys(app) {
		keys[key] = true
	}

	out := make(Messages, len(messages))
	dropped := 0

	for key, value := range messages {
		switch {
		case strings.HasPrefix(key, metaPrefix):
			// The names a file carries are the language's, not the app's;
			// they are read separately and never counted as dropped.
		case !keys[key]:
			dropped++
		case strings.TrimSpace(value) != "":
			out[key] = value
		}
	}

	return out, dropped
}

// catalog is the files, parsed once. They are embedded, so they cannot change
// while the process runs.
type catalog struct {
	files []File
	// keys are each app's keys, sorted, and base the base language's text.
	keys map[App][]string
	base map[App]Messages
}

var load = sync.OnceValues(func() (*catalog, error) {
	byCode := map[string]*File{}

	for _, app := range Apps {
		messages, err := readApp(app)
		if err != nil {
			return nil, err
		}

		for code, file := range messages {
			language := byCode[code]
			if language == nil {
				language = &File{Code: code, Messages: map[App]Messages{}}
				byCode[code] = language
			}

			// A language names itself in whichever of its files has it; the
			// two say the same thing, and a language with only one file still
			// has a name.
			if language.Name == "" {
				language.Name = file["$name"]
			}
			if language.Native == "" {
				language.Native = file["$native"]
			}

			language.Messages[app] = withoutMeta(file)
		}
	}

	shipped := &catalog{keys: map[App][]string{}, base: map[App]Messages{}}

	for _, language := range byCode {
		if language.Name == "" {
			language.Name = language.Code
		}
		if language.Native == "" {
			language.Native = language.Name
		}

		shipped.files = append(shipped.files, *language)
	}

	sort.Slice(shipped.files, func(i, j int) bool {
		a, b := shipped.files[i], shipped.files[j]
		if (a.Code == Base) != (b.Code == Base) {
			return a.Code == Base
		}

		return a.Name < b.Name
	})

	base := byCode[Base]
	if base == nil {
		return nil, fmt.Errorf("there is no %s.json: it is the language every other falls back to", Base)
	}

	for _, app := range Apps {
		shipped.base[app] = base.Messages[app]

		keys := make([]string, 0, len(base.Messages[app]))
		for key := range base.Messages[app] {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		shipped.keys[app] = keys
	}

	return shipped, nil
})

// readApp reads every file for one app, by language code.
func readApp(app App) (map[string]Messages, error) {
	entries, err := files.ReadDir(string(app))
	if err != nil {
		return nil, fmt.Errorf("read %s locales: %w", app, err)
	}

	byCode := make(map[string]Messages, len(entries))

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}

		raw, err := files.ReadFile(path.Join(string(app), name))
		if err != nil {
			return nil, fmt.Errorf("read %s/%s: %w", app, name, err)
		}

		var nested map[string]any
		if err := json.Unmarshal(raw, &nested); err != nil {
			return nil, fmt.Errorf("read %s/%s: %w", app, name, err)
		}

		file, err := Flatten(nested)
		if err != nil {
			return nil, fmt.Errorf("read %s/%s: %w", app, name, err)
		}

		byCode[strings.TrimSuffix(name, ".json")] = file
	}

	return byCode, nil
}

// Flatten reads a translation file into messages by dotted key.
//
// The files are nested by screen, so a translator sees one page's text
// together:
//
//	{"login": {"title": "Sign in", "subtitle_app": "to continue to {app}"}}
//
// and everything else — the database, the editor, the apps' `t()` — speaks
// of "login.title". A file may also be flat, or mix the two, as long as no
// key is said twice; a value that is not text is refused.
func Flatten(nested map[string]any) (Messages, error) {
	out := Messages{}
	if err := flattenInto(out, "", nested); err != nil {
		return nil, err
	}

	return out, nil
}

func flattenInto(out Messages, prefix string, node map[string]any) error {
	for key, value := range node {
		full := key
		if prefix != "" {
			full = prefix + "." + key
		}

		switch value := value.(type) {
		case string:
			if _, twice := out[full]; twice {
				return fmt.Errorf("%s is given twice", full)
			}
			out[full] = value
		case map[string]any:
			if err := flattenInto(out, full, value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%s is not text", full)
		}
	}

	return nil
}

// withoutMeta is a file's messages: everything that is not about the file.
func withoutMeta(file Messages) Messages {
	out := make(Messages, len(file))

	for key, value := range file {
		if !strings.HasPrefix(key, metaPrefix) {
			out[key] = value
		}
	}

	return out
}
