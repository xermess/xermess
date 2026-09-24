package otp

import (
	"net/http"

	"xermess/internal/api/respond"
	"xermess/internal/api/validate"
	"xermess/internal/model"
)

// applyTo copies what a request sent onto the settings and returns what it
// changed, in the order listed here, so the activity log can say what an
// administrator touched.
//
// The bounds are the model's — what a code may be is the same question
// wherever it is asked — and nothing is written when any of them is broken,
// since the record is only saved once this returns.
func (r *settingsRequest) applyTo(settings *model.OTPSettings) ([]string, error) {
	updated := *settings
	updated.Length = validate.Number(r.Length, settings.Length)
	updated.LifetimeMinutes = validate.Number(r.LifetimeMinutes, settings.LifetimeMinutes)
	updated.MaxAttempts = validate.Number(r.MaxAttempts, settings.MaxAttempts)
	updated.ResendSeconds = validate.Number(r.ResendSeconds, settings.ResendSeconds)

	if err := updated.Validate(); err != nil {
		return nil, respond.Fault{Status: http.StatusBadRequest, Message: err.Error()}
	}

	changed := []string{}
	for _, field := range []struct {
		name   string
		was    int
		became int
	}{
		{"length", settings.Length, updated.Length},
		{"lifetime_minutes", settings.LifetimeMinutes, updated.LifetimeMinutes},
		{"max_attempts", settings.MaxAttempts, updated.MaxAttempts},
		{"resend_seconds", settings.ResendSeconds, updated.ResendSeconds},
	} {
		if field.was != field.became {
			changed = append(changed, field.name)
		}
	}

	*settings = updated

	return changed, nil
}
