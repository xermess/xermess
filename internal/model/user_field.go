package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FieldType is the kind of value a user field holds. Adding one here means
// teaching Normalise below how to read it, and the panel how to render and
// edit it.
type FieldType string

const (
	FieldText   FieldType = "text"
	FieldNumber FieldType = "number"
	FieldBool   FieldType = "bool"
	FieldEmail  FieldType = "email"
	FieldDate   FieldType = "date"
)

// FieldTypes lists every type, for validation and for the panel's type picker.
var FieldTypes = []FieldType{FieldText, FieldNumber, FieldBool, FieldEmail, FieldDate}

// Valid reports whether t is a known type.
func (t FieldType) Valid() bool {
	for _, known := range FieldTypes {
		if t == known {
			return true
		}
	}
	return false
}

// UserField describes one field of a user record.
//
// A row of this table is an additional field: something an organisation added
// in the panel, whose values live in users.data, so adding one is a row here
// rather than a migration somebody has to write.
//
// The built-in fields are described with this same struct — see BuiltinFields
// at the bottom of the file — but no row of them exists or ever should: a
// built-in field IS a column of users. Builtin says which of the two you are
// holding.
type UserField struct {
	Base

	Name       string    `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Label      string    `gorm:"size:100;not null" json:"label"`
	Type       FieldType `gorm:"type:varchar(16);not null" json:"type"`
	IsRequired bool      `gorm:"not null;default:false" json:"is_required"`

	// IsUnique means no two users may hold the same value. It is checked when a
	// user is written, since the values live in a JSON column rather than in
	// a column the database could index. The column is is_unique because
	// UNIQUE is a word of SQL's own.
	IsUnique bool `gorm:"not null;default:false" json:"is_unique"`

	// Min and Max bound a number's value, or the length of text. Nil means
	// the field is not bounded that way.
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`

	// StartsWith is a prefix text has to begin with, such as "+" for a phone
	// number an organisation adds as a field of its own. Empty means anything
	// is accepted.
	StartsWith string `gorm:"size:64;not null;default:''" json:"starts_with"`

	// Position orders the columns in the table and the inputs in the form.
	// Built-in fields come first, in the order BuiltinFields lists them.
	Position int `gorm:"not null;default:0" json:"position"`

	// IsBuiltin marks a field that is a column of the record rather than a row
	// of this table. It is never stored — nothing in user_fields is built in
	// — and is here so the panel can be given one list of fields and still
	// know which of them it may edit.
	IsBuiltin bool `gorm:"-" json:"is_builtin"`
}

// Bounded reports whether this type of field can carry a Min and a Max.
// Numbers are bounded by value, text by length; the rest are not bounded.
func (t FieldType) Bounded() bool {
	return t == FieldNumber || t == FieldText || t == FieldEmail
}

// Prefixed reports whether this type of field can carry a StartsWith.
func (t FieldType) Prefixed() bool {
	return t == FieldText || t == FieldEmail
}

// Validate checks the definition itself, as opposed to a value stored under
// it: a bound the field cannot have, or a maximum below its minimum, is a
// mistake in the panel rather than in someone's record.
func (f UserField) Validate() error {
	if !f.Type.Valid() {
		return ErrFieldValue{f.Name, "is not a field type"}
	}

	if !f.Type.Bounded() && (f.Min != nil || f.Max != nil) {
		return ErrFieldValue{f.Name, "cannot have a smallest or largest value"}
	}

	if !f.Type.Prefixed() && f.StartsWith != "" {
		return ErrFieldValue{f.Name, "cannot have a prefix"}
	}

	if f.Min != nil && f.Max != nil && *f.Min > *f.Max {
		return ErrFieldValue{f.Name, "has a largest value below its smallest"}
	}

	if f.Type != FieldNumber && (negative(f.Min) || negative(f.Max)) {
		return ErrFieldValue{f.Name, "is measured in characters, which cannot be negative"}
	}

	return nil
}

func negative(value *float64) bool {
	return value != nil && *value < 0
}

// TableName pins the table name.
func (UserField) TableName() string {
	return "user_fields"
}

// ErrFieldValue is returned when a value does not suit its field. The message
// is shown to whoever is editing, so it says what was expected.
type ErrFieldValue struct {
	Field  string
	Reason string
}

func (e ErrFieldValue) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Reason)
}

// Normalise checks a value against the field and returns it in the form that
// should be stored: numbers as float64, booleans as bool, dates as RFC 3339,
// and everything else trimmed. An empty value comes back as nil, which is how
// "not set" is stored.
func (f UserField) Normalise(value any) (any, error) {
	if value == nil {
		return nil, f.requiredError()
	}

	switch f.Type {
	case FieldBool:
		switch typed := value.(type) {
		case bool:
			return typed, nil
		case string:
			parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
			if err != nil {
				return nil, ErrFieldValue{f.Name, "must be true or false"}
			}
			return parsed, nil
		default:
			return nil, ErrFieldValue{f.Name, "must be true or false"}
		}

	case FieldNumber:
		switch typed := value.(type) {
		case float64:
			return typed, f.checkNumber(typed)
		case int:
			return float64(typed), f.checkNumber(float64(typed))
		case string:
			trimmed := strings.TrimSpace(typed)
			if trimmed == "" {
				return nil, f.requiredError()
			}
			parsed, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, ErrFieldValue{f.Name, "must be a number"}
			}
			return parsed, f.checkNumber(parsed)
		default:
			return nil, ErrFieldValue{f.Name, "must be a number"}
		}

	case FieldDate:
		text, ok := value.(string)
		if !ok {
			return nil, ErrFieldValue{f.Name, "must be a date"}
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			return nil, f.requiredError()
		}
		// The panel sends a plain date; both that and a full timestamp are
		// accepted, and both are stored as a timestamp.
		for _, layout := range []string{time.RFC3339, "2006-01-02"} {
			if parsed, err := time.Parse(layout, trimmed); err == nil {
				return parsed.Format(time.RFC3339), nil
			}
		}
		return nil, ErrFieldValue{f.Name, "must be a date like 2026-09-13"}

	case FieldEmail:
		text, ok := value.(string)
		if !ok {
			return nil, ErrFieldValue{f.Name, "must be an email address"}
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			return nil, f.requiredError()
		}
		if !strings.Contains(trimmed, "@") || strings.HasPrefix(trimmed, "@") || strings.HasSuffix(trimmed, "@") {
			return nil, ErrFieldValue{f.Name, "must be an email address"}
		}
		return trimmed, f.checkText(trimmed)

	default: // FieldText
		text, ok := value.(string)
		if !ok {
			return nil, ErrFieldValue{f.Name, "must be text"}
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			return nil, f.requiredError()
		}
		return trimmed, f.checkText(trimmed)
	}
}

// checkNumber holds a number to the bounds the field was given.
func (f UserField) checkNumber(value float64) error {
	if f.Min != nil && value < *f.Min {
		return ErrFieldValue{f.Name, fmt.Sprintf("must be %s or more", number(*f.Min))}
	}

	if f.Max != nil && value > *f.Max {
		return ErrFieldValue{f.Name, fmt.Sprintf("must be %s or less", number(*f.Max))}
	}

	return nil
}

// checkText holds text to its length bounds and its prefix. Length is counted
// in characters rather than bytes, so an accented name counts as one each.
func (f UserField) checkText(value string) error {
	length := float64(len([]rune(value)))

	if f.Min != nil && length < *f.Min {
		return ErrFieldValue{f.Name, fmt.Sprintf("must be at least %s characters", number(*f.Min))}
	}

	if f.Max != nil && length > *f.Max {
		return ErrFieldValue{f.Name, fmt.Sprintf("must be at most %s characters", number(*f.Max))}
	}

	if f.StartsWith != "" && !strings.HasPrefix(value, f.StartsWith) {
		return ErrFieldValue{f.Name, fmt.Sprintf("must start with %q", f.StartsWith)}
	}

	return nil
}

// number prints a bound the way it was written: 3 rather than 3.0.
func number(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// requiredError is nil for an optional field, so an empty value is simply not
// stored.
func (f UserField) requiredError() error {
	if f.IsRequired {
		return ErrFieldValue{f.Name, "is required"}
	}
	return nil
}

// ---------------------------------------------------------------------------
// The built-in fields: the columns of users, described.
// ---------------------------------------------------------------------------

// The built-in fields of a user record, by name. They are columns of the
// users table — nothing here is ever inserted anywhere — so these names are
// reserved: an additional field may not take one.
const (
	FieldEmailName           = "email"
	FieldIsEmailVerifiedName = "is_email_verified"
	FieldFirstNameName       = "first_name"
	FieldLastNameName        = "last_name"
	FieldIsActiveName        = "is_active"
	// FieldIsPasswordTemporaryName is shown and filtered like the others, but
	// it is set together with the password rather than on its own.
	FieldIsPasswordTemporaryName = "is_password_temporary"
)

// BuiltinFields describes the columns every user record has.
//
// They are described with the same shape as the additional ones so the panel
// can draw one table and one form from a single list, and so the rules a
// value keeps read the same either way. Nothing here is stored in
// user_fields: this is the description of columns that already exist, which
// is why none of them can be edited or removed.
func BuiltinFields() []UserField {
	max := func(n float64) *float64 { return &n }

	return []UserField{
		{
			Name:       FieldEmailName,
			Label:      "Email",
			Type:       FieldEmail,
			IsRequired: true,
			IsUnique:   true,
			Max:        max(255),
			IsBuiltin:  true,
			Position:   1,
		},
		{
			Name:      FieldIsEmailVerifiedName,
			Label:     "Is Email Verified",
			Type:      FieldBool,
			IsBuiltin: true,
			Position:  2,
		},
		{
			Name:      FieldFirstNameName,
			Label:     "First name",
			Type:      FieldText,
			Max:       max(100),
			IsBuiltin: true,
			Position:  3,
		},
		{
			Name:      FieldLastNameName,
			Label:     "Last name",
			Type:      FieldText,
			Max:       max(100),
			IsBuiltin: true,
			Position:  4,
		},
		{
			Name:      FieldIsActiveName,
			Label:     "Is Active",
			Type:      FieldBool,
			IsBuiltin: true,
			Position:  5,
		},
		{
			Name:      FieldIsPasswordTemporaryName,
			Label:     "Is Temporary",
			Type:      FieldBool,
			IsBuiltin: true,
			Position:  6,
		},
	}
}

// IsBuiltinField reports whether a name belongs to a column of the record.
// Additional fields are checked against this: two fields with one name would
// be two places to look for the same thing.
func IsBuiltinField(name string) bool {
	for _, field := range BuiltinFields() {
		if field.Name == name {
			return true
		}
	}

	return false
}
