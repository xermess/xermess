// Package locales is the translations this server ships with, and the one
// list of which keys there are.
//
// A shipped language is a JSON file: `locales/id/ky.json` is Kyrgyz for the
// sign-in pages, `locales/console/ky.json` is Kyrgyz for the admin panel.
// They are embedded, so a deployment is one binary and carries them.
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

		var file Messages
		if err := json.Unmarshal(raw, &file); err != nil {
			return nil, fmt.Errorf("read %s/%s: %w", app, name, err)
		}

		byCode[strings.TrimSuffix(name, ".json")] = file
	}

	return byCode, nil
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
