package store

import (
	"reflect"
	"strings"
	"testing"

	"xermess/internal/model"
)

// What the store caches goes to Redis as JSON, and a field JSON leaves out
// comes back empty. For most fields that would be a quiet bug — a cached
// flow without its steps — and for a secret it would be worse: a copy that
// looks whole and is not. So every field of every cached type has to be one
// JSON writes, and the only exceptions are named here with why.
func TestCachedTypesSurviveJSON(t *testing.T) {
	// Fields JSON may leave out of a cached row.
	allowed := map[string]string{
		"Base.DeletedAt":        "a cached row is a live one, so it is always empty",
		"Language.Translations": "a relation the cached reads never load",
	}

	cachedTypes := []any{
		model.Language{},
		model.Organization{},
		model.LoginFlow{},
		SocialButton{},
	}

	for _, value := range cachedTypes {
		typ := reflect.TypeOf(value)

		t.Run(typ.Name(), func(t *testing.T) {
			for _, path := range fieldsJSONDrops(typ, typ.Name()) {
				if _, ok := allowed[path]; !ok {
					t.Errorf("%s is tagged json:\"-\", so a cached copy would lose it; cache something without it, or say here why that is fine", path)
				}
			}
		})
	}
}

// fieldsJSONDrops names every exported field, embedded ones included, that
// encoding/json would not write.
func fieldsJSONDrops(typ reflect.Type, name string) []string {
	var out []string

	for i := range typ.NumField() {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			out = append(out, fieldsJSONDrops(field.Type, field.Type.Name())...)
			continue
		}

		if tag, _, _ := strings.Cut(field.Tag.Get("json"), ","); tag == "-" {
			out = append(out, name+"."+field.Name)
		}
	}

	return out
}
