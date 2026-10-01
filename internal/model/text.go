package model

import (
	"strings"
	"unicode/utf8"
)

// Truncate makes a string fit a column: invalid UTF-8 is dropped, because
// Postgres refuses it, and the rest is cut to `max` bytes on a rune boundary.
// It is for what arrives from outside and is only recorded — a User-Agent, an
// address typed at a sign-in — where losing the tail is better than losing
// the row: a line of the activity log that Postgres refused is a sign-in
// nobody can see happened.
func Truncate(value string, max int) string {
	value = strings.ToValidUTF8(value, "")
	if len(value) <= max {
		return value
	}

	// Back up off a continuation byte to the start of the rune it belongs to.
	for max > 0 && !utf8.RuneStart(value[max]) {
		max--
	}

	return value[:max]
}
