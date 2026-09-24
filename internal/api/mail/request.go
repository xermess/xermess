package mail

// What this page's records are called in the activity log. There is one of
// each, so the target is the page rather than a row's id.
const (
	targetType = "mail"
	targetID   = "settings"

	// The words of the emails are their own subject, named by the language
	// they were written in: "changed the email content" rather than a second
	// kind of change to the mail settings.
	contentTargetType = "mail_content"
)

// settingsRequest is what an update sends.
//
// Every field is a pointer because this is a PATCH: nil is "not sent", and
// the stored value stands. That matters most for the password — the panel
// never has the stored one to send back, so leaving it out is how a form
// saves without changing it, and an empty string is how it is cleared.
type settingsRequest struct {
	Enabled     *bool   `json:"enabled"`
	Host        *string `json:"host"`
	Port        *int    `json:"port"`
	Encryption  *string `json:"encryption"`
	Username    *string `json:"username"`
	Password    *string `json:"password"`
	FromAddress *string `json:"from_address"`
	FromName    *string `json:"from_name"`
}

// testRequest is a test message: where to send it, and the settings to send
// it with. The settings are the form's, so a server can be tried before it is
// saved; what the form leaves out is the stored value.
type testRequest struct {
	To string `json:"to"`

	settingsRequest
}

// contentRequest is one language's words for the emails: the keys of
// model.MailMessageSpecs and what they say. A key with an empty value clears
// the override rather than storing a blank, and the shipped text comes back.
type contentRequest struct {
	Messages map[string]string `json:"messages"`
}
