package store

import "testing"

// Which columns the database browser will not read is worth testing without a
// database: it is the one rule standing between an administrator who may read
// the tables and the passwords, keys and tokens in them — and it has to be
// wrong in neither direction, since a column hidden by mistake says the table
// holds something it does not.
func TestHiddenColumn(t *testing.T) {
	tests := []struct {
		name   string
		column string
		want   bool
	}{
		{name: "a password", column: "password_hash", want: true},
		{name: "a session token", column: "token_hash", want: true},
		{name: "an authorization code", column: "code_hash", want: true},
		{name: "the handle of a sign-in under way", column: "handle_hash", want: true},
		{name: "the state of a social sign-in", column: "state_hash", want: true},
		{name: "a client secret", column: "client_secret", want: true},
		{name: "an authenticator's secret", column: "secret", want: true},
		{name: "recovery codes", column: "recovery_codes", want: true},
		{name: "a signing key", column: "private_key", want: true},
		{name: "a column named in capitals", column: "Password_Hash", want: true},

		{name: "the last four characters of a secret", column: "secret_hint", want: false},
		{name: "when a secret was made", column: "secret_created_at", want: false},
		{name: "whether a flow offers password resets", column: "allow_password_reset", want: false},
		{name: "whether a password is a temporary one", column: "is_temporary_password", want: false},
		{name: "which key signed something", column: "key_id", want: false},
		{name: "an email address", column: "email", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hiddenColumn(tt.column); got != tt.want {
				t.Errorf("hiddenColumn(%q) = %v, want %v", tt.column, got, tt.want)
			}
		})
	}
}
