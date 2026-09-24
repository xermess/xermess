// Package mail answers the endpoints the Mail page is built on: the server
// this installation sends email through, a test message to try it with, and
// the words of every email the server sends.
//
// The settings are one record, like the organisation's, so there is no list
// and nothing to create or delete. The words are not a record of their own:
// an email goes out in the reader's language, so its subject and body are two
// keys of the sign-in pages' text (model.MailMessageSpecs), and this page
// edits those keys rather than keeping a second copy of them.
package mail

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"xermess/i18n"
	"xermess/internal/api/audit"
	"xermess/internal/api/respond"
	"xermess/internal/jose"
	"xermess/internal/mail"
	"xermess/internal/model"
	"xermess/internal/store"
)

// Handler holds what these endpoints need. The sealer is here because the
// mail server's password is stored sealed, as a social provider's secret is:
// this is the only place that seals one.
type Handler struct {
	store  *store.Store
	sealer *jose.Sealer
	audit  audit.Recorder
	log    *slog.Logger
}

// New returns a Handler.
func New(st *store.Store, sealer *jose.Sealer, recorder audit.Recorder, log *slog.Logger) *Handler {
	return &Handler{store: st, sealer: sealer, audit: recorder, log: log}
}

// Get returns the mail settings.
func (h *Handler) Get(c *gin.Context) {
	settings, err := h.store.MailSettings(c.Request.Context())
	if err != nil {
		respond.Failure(c, h.log, err, "loading the mail settings failed")
		return
	}

	c.JSON(http.StatusOK, newResponse(*settings))
}

// Update changes the settings. What a request leaves out is left as it is, so
// a client that knows about one field does not clear the rest — and the
// password is left as it is unless one was typed, since nothing ever reads
// the stored one back to send it here again.
func (h *Handler) Update(c *gin.Context) {
	settings, err := h.store.MailSettings(c.Request.Context())
	if err != nil {
		respond.Failure(c, h.log, err, "loading the mail settings failed")
		return
	}

	var req settingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	changed, err := req.applyTo(settings, h.sealer)
	if err != nil {
		respond.Failure(c, h.log, err, "validating the mail settings failed")
		return
	}

	if len(changed) == 0 {
		c.JSON(http.StatusOK, newResponse(*settings))
		return
	}

	if err := h.store.SaveMailSettings(c.Request.Context(), settings); err != nil {
		respond.Failure(c, h.log, err, "updating the mail settings failed")
		return
	}

	// Which settings moved, never what to: the password is one of them, and
	// the activity log is read by more people than this page is.
	h.audit.RecordWith(c, "mail.settings_updated", targetType, targetID, map[string]any{
		"fields": changed,
	})

	c.JSON(http.StatusOK, newResponse(*settings))
}

// Test sends one message, to find out whether the settings work before
// anybody's sign-in depends on them.
//
// It sends with what the form holds rather than with what is stored, since
// the point is to try a server before saving it — except for the password,
// which the form only holds when one has just been typed. Left out, the
// stored one is used, so a working server can be tested again without typing
// its password each time.
func (h *Handler) Test(c *gin.Context) {
	var req testRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	stored, err := h.store.MailSettings(c.Request.Context())
	if err != nil {
		respond.Failure(c, h.log, err, "loading the mail settings failed")
		return
	}

	settings, err := req.settings(stored, h.sealer)
	if err != nil {
		respond.Failure(c, h.log, err, "reading the mail settings to test failed")
		return
	}

	to := req.To
	if to == "" {
		respond.Fail(c, testNeedsRecipient)
		return
	}

	err = mail.Deliver(c.Request.Context(), settings, mail.Message{
		To:      to,
		Subject: "xermess: a test message",
		Body:    testBody,
	})
	if err != nil {
		// What the mail server said is the whole point of the test, so it is
		// given back rather than logged and hidden behind a 500.
		h.log.Warn("a test email failed", "error", err, "host", settings.Host)
		h.audit.RecordWith(c, "mail.test_failed", targetType, targetID, map[string]any{
			"to": to, "host": settings.Host,
		})

		respond.Fail(c, testFailed, "reason", err.Error())

		return
	}

	h.audit.RecordWith(c, "mail.test_sent", targetType, targetID, map[string]any{
		"to": to, "host": settings.Host,
	})

	c.JSON(http.StatusOK, gin.H{"sent": true, "to": to})
}

// testBody is what the test message says. It is not a translated email: it is
// read by the administrator who pressed the button, in the panel they pressed
// it in, and it exists to prove a connection rather than to tell anybody
// anything.
const testBody = "This is a test message from your xermess installation.\n\n" +
	"If you are reading it, the mail settings on the Mail page work: " +
	"password resets, address confirmations and one-time codes will reach people.\n"

// Content returns the words of every email the server sends, for each
// language the sign-in pages are offered in: what the language says, and what
// English says underneath it.
func (h *Handler) Content(c *gin.Context) {
	ctx := c.Request.Context()

	languages, err := h.store.Languages(ctx)
	if err != nil {
		respond.Failure(c, h.log, err, "listing languages failed")
		return
	}

	english := i18n.Resolve(i18n.ID)

	out := make([]languageContent, 0, len(languages))
	for _, language := range languages {
		messages, err := h.store.Translation(ctx, language.ID, string(i18n.ID))
		if err != nil {
			respond.Failure(c, h.log, err, "loading a translation failed")
			return
		}

		out = append(out, newLanguageContent(language, messages, english))
	}

	c.JSON(http.StatusOK, contentResponse{Messages: model.MailMessageSpecs, Languages: out})
}

// SaveContent writes one language's words for the emails.
//
// Only the keys the messages are made of: this page holds a handful of a
// language's text and has the rest nowhere, so it merges rather than
// replacing (store.SaveTranslationKeys), and a key that is not an email's is
// refused rather than quietly dropped — a page sending one is a page with a
// bug, not an old file being imported.
func (h *Handler) SaveContent(c *gin.Context) {
	language, err := h.store.Language(c.Request.Context(), c.Param("code"))
	if errors.Is(err, store.ErrNotFound) {
		respond.Fail(c, respond.LanguageNotFound)
		return
	}
	if err != nil {
		respond.Failure(c, h.log, err, "loading a language failed")
		return
	}

	var req contentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	if err := req.validate(); err != nil {
		respond.Failure(c, h.log, err, "validating email content failed")
		return
	}

	if err := h.store.SaveTranslationKeys(c.Request.Context(), language, string(i18n.ID), req.Messages); err != nil {
		respond.Failure(c, h.log, err, "saving email content failed")
		return
	}

	h.audit.RecordWith(c, "mail.content_updated", contentTargetType, language.Code, map[string]any{
		"language": language.Code, "keys": len(req.Messages),
	})

	messages, err := h.store.Translation(c.Request.Context(), language.ID, string(i18n.ID))
	if err != nil {
		respond.Failure(c, h.log, err, "loading a translation failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"language": newLanguageContent(*language, messages, i18n.Resolve(i18n.ID)),
	})
}
