package otp

// What this page's record is called in the activity log. There is one of it,
// so the target is the page rather than a row's id.
const (
	targetType = "otp"
	targetID   = "settings"
)

// settingsRequest is what an update sends.
//
// Every field is a pointer because this is a PATCH: nil is "not sent", and
// the stored value stands.
type settingsRequest struct {
	Length          *int `json:"length"`
	LifetimeMinutes *int `json:"lifetime_minutes"`
	MaxAttempts     *int `json:"max_attempts"`
	ResendSeconds   *int `json:"resend_seconds"`
}
