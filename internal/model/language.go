package model

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Language is a language the sign-in pages may offer, with its text in
// Translations, one row per app. Shipped languages are imported on the first
// start; after that this table is the list. Code is an IETF tag such as "ky" or
// "pt-BR".
type Language struct {
	Base

	Code string `gorm:"size:16;not null;uniqueIndex" json:"code"`

	// Name is the English name for the panel; NativeName is what the language
	// calls itself, shown in the picker.
	Name       string `gorm:"size:64;not null" json:"name"`
	NativeName string `gorm:"size:64;not null" json:"native_name"`

	// IsEnabled shows the language in the picker; the default language is
	// always enabled.
	IsEnabled bool `gorm:"not null" json:"is_enabled"`

	// IsDefault marks the language shown before anyone chooses; exactly one row
	// has it.
	IsDefault bool `gorm:"not null" json:"is_default"`

	// Position is where it comes in the picker, lowest first.
	Position int `gorm:"not null" json:"position"`

	Translations []Translation `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// TableName pins the table name.
func (Language) TableName() string {
	return "languages"
}

// BaseLanguage defines the keys and the fallback text. It cannot be removed.
const BaseLanguage = "en"

// DefaultLanguage is the row a fresh installation starts with.
func DefaultLanguage() Language {
	return Language{
		Code: BaseLanguage, Name: "English", NativeName: "English",
		IsEnabled: true, IsDefault: true, Position: 1,
	}
}

// languageCodePattern is a narrow IETF tag (language, optional script or
// region), since codes go into URLs, cookies and file names.
var languageCodePattern = regexp.MustCompile(`^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

// Validate reports the first thing wrong with a language.
func (l Language) Validate() error {
	switch {
	case strings.TrimSpace(l.Code) == "":
		return fmt.Errorf("code is required")
	case len(l.Code) > 16:
		return fmt.Errorf("code must be at most 16 characters")
	case !languageCodePattern.MatchString(l.Code):
		return fmt.Errorf("code must be a language tag, such as en, ky or pt-BR")
	case strings.TrimSpace(l.Name) == "":
		return fmt.Errorf("name is required")
	case utf8.RuneCountInString(l.Name) > 64:
		return fmt.Errorf("name must be at most 64 characters")
	case strings.TrimSpace(l.NativeName) == "":
		return fmt.Errorf("native name is required")
	case utf8.RuneCountInString(l.NativeName) > 64:
		return fmt.Errorf("native name must be at most 64 characters")
	case l.Position < 0:
		return fmt.Errorf("position cannot be negative")
	}

	// The default is what somebody sees before choosing, so it has to be
	// among what they could choose.
	if l.IsDefault && !l.IsEnabled {
		return fmt.Errorf("the default language cannot be turned off")
	}

	return nil
}

// Translation is one language's text for one app, without the file's `$name`
// and `$native`.
type Translation struct {
	Base

	LanguageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_translations_language_app,priority:1" json:"language_id"`
	App        string    `gorm:"size:16;not null;uniqueIndex:idx_translations_language_app,priority:2" json:"app"`

	// Messages is the text by key. A key that is not here is not translated,
	// and the app is sent the base language's text for it instead.
	Messages map[string]string `gorm:"type:jsonb;serializer:json;not null" json:"messages"`
}

// TableName pins the table name.
func (Translation) TableName() string {
	return "translations"
}

// MaxMessageLength leaves room for a long paragraph, not a document.
const MaxMessageLength = 2000

// TooLongMessage returns the key of a message over MaxMessageLength, or "".
func TooLongMessage(messages map[string]string) string {
	for key, value := range messages {
		if utf8.RuneCountInString(value) > MaxMessageLength {
			return key
		}
	}

	return ""
}
