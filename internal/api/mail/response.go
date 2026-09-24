package mail

import (
	"net/http"

	"xermess/internal/api/respond"
	"xermess/internal/model"
)

// The problems these endpoints answer with.
var (
	// testNeedsRecipient is pressing "send a test" with nowhere to send it.
	testNeedsRecipient = respond.Define(http.StatusBadRequest, "mail_test_needs_recipient", respond.Admin)

	// testFailed is the mail server refusing or not answering. What it said
	// is the `reason`, since finding that out is the whole point of a test.
	testFailed = respond.Define(http.StatusBadGateway, "mail_test_failed", respond.Admin)

	// contentKeyUnknown is a key that is not one of an email's, which the
	// Mail page has no business writing.
	contentKeyUnknown = respond.Define(http.StatusBadRequest, "mail_content_key_unknown", respond.Admin)
)

// settingsResponse is the mail settings as the panel sees them. It is built
// by hand rather than returning the model so the row's own bookkeeping stays
// out of a record of settings — and so the password cannot leave by accident:
// what is said about it is whether there is one.
type settingsResponse struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Encryption  string `json:"encryption"`
	Username    string `json:"username"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`

	// HasPassword says a password is stored, without saying what it is. The
	// form shows "set" rather than asking for one that is already there.
	HasPassword bool `json:"has_password"`
}

// response is what the two settings endpoints answer with. The encryptions
// come along so the panel builds its picker from the server's catalog rather
// than from a list of its own.
type response struct {
	Mail        settingsResponse       `json:"mail"`
	Encryptions []model.MailEncryption `json:"encryptions"`
}

func newResponse(settings model.MailSettings) response {
	return response{
		Mail: settingsResponse{
			Enabled:     settings.Enabled,
			Host:        settings.Host,
			Port:        settings.Port,
			Encryption:  string(settings.Encryption),
			Username:    settings.Username,
			FromAddress: settings.FromAddress,
			FromName:    settings.FromName,
			HasPassword: len(settings.Password) > 0,
		},
		Encryptions: model.MailEncryptions,
	}
}

// languageContent is one language's words for the emails: what it says, and
// what English says, key by key.
type languageContent struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Native string `json:"native"`

	// Offered is whether the sign-in pages are shown in this language. One
	// that is not still has its text, and the panel marks it.
	Offered bool `json:"offered"`

	// Messages is what this language says, for the email keys alone, and
	// only where it says something: a key it is missing falls back to Base.
	Messages map[string]string `json:"messages"`

	// Base is the shipped English for the same keys, which is what an email
	// goes out in where a language has nothing.
	Base map[string]string `json:"base"`
}

func newLanguageContent(language model.Language, messages, english map[string]string) languageContent {
	own := map[string]string{}
	base := map[string]string{}

	for _, key := range model.MailMessageKeys() {
		if text, ok := messages[key]; ok {
			own[key] = text
		}
		base[key] = english[key]
	}

	return languageContent{
		Code:     language.Code,
		Name:     language.Name,
		Native:   language.Native,
		Offered:  language.Enabled,
		Messages: own,
		Base:     base,
	}
}

// contentResponse is every email the server sends and every language it can
// be sent in: the catalog once, and the text per language.
type contentResponse struct {
	Messages  []model.MailMessageSpec `json:"messages"`
	Languages []languageContent       `json:"languages"`
}
