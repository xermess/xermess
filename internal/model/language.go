package model

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Language is one language this installation has: what it is called, whether
// the sign-in pages offer it, whether it is the default, and where it comes in
// the picker. Its text is in Translations, one row per app.
//
// The languages the server ships with (locales/) are imported on the first
// start, and from then on this table is the list: a language an administrator
// adds exists only here, and one they remove is gone, shipped or not.
//
// The code is an IETF language tag, "ky" or "pt-BR".
type Language struct {
	Base

	Code string `gorm:"size:16;not null;uniqueIndex" json:"code"`

	// Name is what the language is called in English, for the panel, and
	// Native what it calls itself — which is what the picker on the sign-in
	// pages shows, since somebody looking for their language scans for it
	// written in it.
	Name   string `gorm:"size:64;not null" json:"name"`
	Native string `gorm:"size:64;not null" json:"native"`

	// Enabled shows the language in the picker on the sign-in pages. The
	// default language is always enabled — it is what somebody sees before
	// choosing, so it has to be one they could choose.
	Enabled bool `gorm:"not null" json:"enabled"`

	// IsDefault marks the language somebody sees before they have chosen one.
	// Exactly one row has it; the store moves it rather than letting two rows
	// claim it.
	IsDefault bool `gorm:"not null" json:"is_default"`

	// Position is where it comes in the picker, lowest first.
	Position int `gorm:"not null" json:"position"`

	Translations []Translation `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// TableName pins the table name.
func (Language) TableName() string {
	return "languages"
}

// BaseLanguage is the language every other is a translation of: the keys it
// has are the keys the apps look up, and its text is what a missing
// translation falls back to. It cannot be removed. It is spelled the same
// here as in locales/.
const BaseLanguage = "en"

// DefaultLanguage is the row a fresh installation starts with.
func DefaultLanguage() Language {
	return Language{
		Code: BaseLanguage, Name: "English", Native: "English",
		Enabled: true, IsDefault: true, Position: 1,
	}
}

// languageCodePattern is an IETF language tag as this project uses them: a
// two- or three-letter language, optionally a script or a region after a
// dash. It is deliberately narrow — the code goes into URLs, cookies and
// file names.
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
	case strings.TrimSpace(l.Native) == "":
		return fmt.Errorf("native name is required")
	case utf8.RuneCountInString(l.Native) > 64:
		return fmt.Errorf("native name must be at most 64 characters")
	case l.Position < 0:
		return fmt.Errorf("position cannot be negative")
	}

	// The default is what somebody sees before choosing, so it has to be
	// among what they could choose.
	if l.IsDefault && !l.Enabled {
		return fmt.Errorf("the default language cannot be turned off")
	}

	return nil
}

// Translation is one language's text for one app: the sign-in pages ("id") or
// the admin panel ("console"). It is the JSON a translator writes, minus the
// `$name` and `$native` that describe the file — those are the language's own
// columns.
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

// MaxMessageLength is the longest one message may be. The longest the apps
// have is a paragraph; this is room for a translation that runs long, not for
// a document.
const MaxMessageLength = 2000

// ValidateMessages reports the first message that is too long. Which keys are
// allowed is not the model's to know — that is the base language's files.
func ValidateMessages(messages map[string]string) error {
	for key, value := range messages {
		if utf8.RuneCountInString(value) > MaxMessageLength {
			return fmt.Errorf("the text of %s must be at most %d characters", key, MaxMessageLength)
		}
	}

	return nil
}
