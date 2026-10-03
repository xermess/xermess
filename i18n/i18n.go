// Package i18n holds the shipped translations and the list of keys that exist.
//
// A language is a directory of JSON groups (`i18n/id/ru/` is Russian for the
// sign-in pages), nested by namespace and flattened to dotted keys such as
// "login.title". Only the sign-in pages are translated; `i18n/console/en/`
// holds just the English sentences for the admin API's problems.
//
// What an installation serves lives in the database: the first start imports
// everything, and afterwards the shipped files only matter as the base
// language's key list (and last-resort fallback text) and as the source of keys
// a release adds.
package i18n

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

//go:embed id/*/*.json console/*/*.json
var files embed.FS

// App is one of the two apps a translation is for.
type App string

const (
	// ID is the sign-in pages and a user's own account, web/id.
	ID App = "id"

	// Console is the admin panel. It is not translated; it is an app here only
	// so the English of its problems has a home the tests can check.
	Console App = "console"
)

// Apps are the apps a language is translated for, in the order the panel
// lists them. The panel is not one of them: see the note on Console.
var Apps = []App{ID}

// catalogued are the apps there are shipped text for at all — the ones a
// language is translated for, and the admin API's English beside them.
var catalogued = []App{ID, Console}

// ParseApp reads an app's name, as it appears in a path.
func ParseApp(name string) (App, bool) {
	for _, app := range Apps {
		if string(app) == name {
			return app, true
		}
	}

	return "", false
}

// AppsFor returns the apps a language is translated for: only the sign-in
// pages.
func AppsFor(string) []App {
	return Apps
}

// ServesApp reports whether a language is translated for an app.
func ServesApp(code string, app App) bool {
	return slices.Contains(AppsFor(code), app)
}

// Base is the language every other translates; its keys are the keys there are.
const Base = "en"

// Messages is one language's text for one app, by key.
type Messages = map[string]string

// File is one language this server ships with, after its groups are merged.
type File struct {
	// Code is the language directory's name: an IETF language tag, "ky" or
	// "pt-BR".
	Code string

	// Name and Native are the language's English and native names, from its
	// metadata.
	Name   string
	Native string

	// Messages is the text for each app, without the `$` keys that describe
	// the language. An app the language has no groups for is left out.
	Messages map[App]Messages
}

// metaPrefix marks the keys that are about the language rather than in the
// interface: they are not counted, and never looked up.
const metaPrefix = "$"

// Shipped is every language there are groups for, the base language first and
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

// Keys are an app's keys from the base language, sorted (which groups each
// namespace).
func Keys(app App) []string {
	shipped, err := load()
	if err != nil {
		return nil
	}

	return shipped.keys[app]
}

// Coverage is the percentage of an app's keys a language translates and how
// many are missing; blank values do not count.
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

// Resolve returns every key for one language, taken from the first of `layers`
// that has it, else the shipped base language, so a page never shows a bare
// key. Keys not in the base language are dropped.
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

// Text is the shipped base-language text for a key: the server's own English,
// such as an error's sentence.
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

// Known keeps only keys the app looks up, dropping empty values, and returns
// how many it dropped.
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
			// The names a catalog carries are the language's, not the app's;
			// they are read separately and never counted as dropped.
		case !keys[key]:
			dropped++
		case strings.TrimSpace(value) != "":
			out[key] = value
		}
	}

	return out, dropped
}

// catalog is the merged shipped text, parsed once. They are embedded, so they
// cannot change while the process runs.
type catalog struct {
	files []File
	// keys are each app's keys, sorted, and base the base language's text.
	keys map[App][]string
	base map[App]Messages
}

var load = sync.OnceValues(func() (*catalog, error) {
	byCode := map[string]*File{}

	// Only translatable apps make languages; the admin API's English goes
	// straight into the base text.
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

			// A language may name itself in any of its groups.
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
		return nil, fmt.Errorf("there is no %s translation directory: it is the language every other falls back to", Base)
	}

	for _, app := range catalogued {
		messages := base.Messages[app]

		// The apps a language is translated for have their base text in the
		// base language's catalog already; the rest is read on its own.
		if messages == nil {
			files, err := readApp(app)
			if err != nil {
				return nil, err
			}
			messages = withoutMeta(files[Base])
		}

		shipped.base[app] = messages

		keys := make([]string, 0, len(messages))
		for key := range messages {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		shipped.keys[app] = keys
	}

	return shipped, nil
})

// readApp reads every language directory for one app and merges its groups.
func readApp(app App) (map[string]Messages, error) {
	entries, err := files.ReadDir(string(app))
	if err != nil {
		return nil, fmt.Errorf("read %s translations: %w", app, err)
	}

	byCode := make(map[string]Messages, len(entries))

	for _, languageEntry := range entries {
		if !languageEntry.IsDir() {
			continue
		}

		code := languageEntry.Name()
		groups, err := files.ReadDir(path.Join(string(app), code))
		if err != nil {
			return nil, fmt.Errorf("read %s/%s translations: %w", app, code, err)
		}

		messages := Messages{}
		for _, group := range groups {
			name := group.Name()
			if group.IsDir() || !strings.HasSuffix(name, ".json") {
				continue
			}

			groupPath := path.Join(string(app), code, name)
			raw, err := files.ReadFile(groupPath)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", groupPath, err)
			}

			var nested map[string]any
			if err := json.Unmarshal(raw, &nested); err != nil {
				return nil, fmt.Errorf("read %s: %w", groupPath, err)
			}

			file, err := Flatten(nested)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", groupPath, err)
			}

			for key, value := range file {
				if _, twice := messages[key]; twice {
					return nil, fmt.Errorf("read %s: %s is given twice", groupPath, key)
				}
				messages[key] = value
			}
		}

		if len(messages) > 0 {
			byCode[code] = messages
		}
	}

	return byCode, nil
}

// Flatten turns a nested translation group into dotted keys:
//
//	{"login": {"title": "Sign in"}}  ->  "login.title"
//
// Flat and mixed groups are accepted as long as no key repeats; non-text values
// are refused.
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

// withoutMeta is a language's messages: everything that is not about the
// language itself.
func withoutMeta(file Messages) Messages {
	out := make(Messages, len(file))

	for key, value := range file {
		if !strings.HasPrefix(key, metaPrefix) {
			out[key] = value
		}
	}

	return out
}
