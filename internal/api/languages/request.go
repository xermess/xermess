package languages

// targetType is what a language is called in the activity log.
const targetType = "language"

// createRequest is what adding a language sends.
//
// CopyFrom names where its text starts: another language here, whose text is
// copied — Portuguese for Brazilian Portuguese — or a language the server
// ships with that this installation has not got, which brings its shipped
// text back. Empty is a language with nothing translated yet, whose pages
// show the base language's text until somebody writes its own.
type createRequest struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Native    string `json:"native"`
	Enabled   *bool  `json:"enabled"`
	IsDefault *bool  `json:"is_default"`
	CopyFrom  string `json:"copy_from"`
}

// languageRequest is what an update sends.
//
// Every field is a pointer because this is a PATCH: nil is "not sent", and
// the stored value stands. The code is not among them — it is in the path,
// and it is what every reader's saved choice names.
type languageRequest struct {
	Name      *string `json:"name"`
	Native    *string `json:"native"`
	Enabled   *bool   `json:"enabled"`
	IsDefault *bool   `json:"is_default"`
	Position  *int    `json:"position"`
}

// translationRequest is one language's whole text for one app, replacing
// what was there: what the editor saves, and what an imported file is.
//
// `$name` and `$native` may come along — a file a translator wrote carries
// them — and are ignored here: the names are the language's own settings.
type translationRequest struct {
	Messages map[string]string `json:"messages"`
}
