// Package languages answers the endpoints behind the panel's Languages page,
// and the two the panel itself is drawn with.
//
// A language is a row and its text, one JSON object per app, all in the
// database. The server ships some (i18n/) and imports them on its first
// start; from then on this is where languages are added, reworded, offered,
// made the default and removed, and a change is on the next page anybody
// opens — the apps ask for their text while rendering, so nothing is rebuilt.
//
// Which keys there are is not the database's to say. It is the base
// language's shipped groups: they are what the apps' code looks up, so a
// translation is counted against them and a key they do not have is dropped.
package languages

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"loginer/i18n"
	"loginer/internal/api/audit"
	"loginer/internal/api/respond"
	"loginer/internal/api/validate"
	"loginer/internal/model"
	"loginer/internal/store"
)

// Handler holds what these endpoints need.
type Handler struct {
	store *store.Store
	audit audit.Recorder
	log   *slog.Logger
}

// New returns a Handler.
func New(st *store.Store, recorder audit.Recorder, log *slog.Logger) *Handler {
	return &Handler{store: st, audit: recorder, log: log}
}

// List returns every language, with how much of each app it translates, and
// the shipped languages this installation does not have.
func (h *Handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	// Asking for the default first creates the row on an installation that
	// has none, so the page is never without the language everything else
	// falls back to.
	if _, err := h.store.DefaultLanguage(ctx); err != nil {
		respond.Failure(c, h.log, err, "loading the default language failed")
		return
	}

	languages, err := h.store.Languages(ctx)
	if err != nil {
		respond.Failure(c, h.log, err, "listing languages failed")
		return
	}

	ids := make([]uuid.UUID, 0, len(languages))
	for _, language := range languages {
		ids = append(ids, language.ID)
	}

	text, err := h.store.Translations(ctx, ids...)
	if err != nil {
		respond.Failure(c, h.log, err, "loading translations failed")
		return
	}

	shipped, err := i18n.Shipped()
	if err != nil {
		respond.Failure(c, h.log, err, "reading the shipped languages failed")
		return
	}

	c.JSON(http.StatusOK, newListResponse(languages, text, shipped))
}

// Create adds a language, with the text it starts from (createRequest).
func (h *Handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	text, file, err := h.startingText(c, strings.TrimSpace(req.CopyFrom))
	if err != nil {
		respond.Failure(c, h.log, err, "loading the text a language starts from failed")
		return
	}

	// Bringing back a shipped language needs no names typed in: the file
	// has them.
	if file != nil && normalizeCode(req.Code) == file.Code {
		req.Name = firstOf(req.Name, file.Name)
		req.Native = firstOf(req.Native, file.Native)
	}

	language, err := req.newLanguage()
	if err != nil {
		respond.Failure(c, h.log, err, "validating a language failed")
		return
	}

	// A copy brings the text of the apps a language is translated for, and
	// nothing else: an older installation's row for an app there no longer
	// is would be copied forward for ever otherwise.
	for app := range text {
		if !i18n.ServesApp(language.Code, i18n.App(app)) {
			delete(text, app)
		}
	}

	language.Position = h.store.NextLanguagePosition(ctx)

	if err := h.store.CreateLanguage(ctx, language, text); err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			respond.Fail(c, codeTaken)
			return
		}

		respond.Failure(c, h.log, err, "creating a language failed")
		return
	}

	h.audit.RecordWith(c, "language.created", targetType, language.ID.String(), map[string]any{
		"language": language.Code, "copied_from": req.CopyFrom,
	})

	h.answer(c, http.StatusCreated, language)
}

// startingText is the text a new language copies, by app: another language's
// here, or a shipped language's when there is no such language here. The file
// is returned too when that is where the text came from.
func (h *Handler) startingText(c *gin.Context, from string) (map[string]map[string]string, *i18n.File, error) {
	if from == "" {
		return nil, nil, nil
	}

	source, err := h.store.Language(c.Request.Context(), from)
	switch {
	case err == nil:
		text, err := h.store.Translations(c.Request.Context(), source.ID)
		return text[source.ID], nil, err
	case !errors.Is(err, store.ErrNotFound):
		return nil, nil, err
	}

	shipped, err := i18n.Shipped()
	if err != nil {
		return nil, nil, err
	}

	for _, file := range shipped {
		if file.Code != from {
			continue
		}

		text := make(map[string]map[string]string, len(file.Messages))
		for app, messages := range file.Messages {
			text[string(app)] = messages
		}

		return text, &file, nil
	}

	return nil, nil, nothingToCopy.With("code", from)
}

// Update changes a language's names and whether and where it is offered.
func (h *Handler) Update(c *gin.Context) {
	language, ok := h.find(c)
	if !ok {
		return
	}

	var req languageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}

	wasDefault := language.IsDefault

	if err := req.applyTo(language); err != nil {
		respond.Failure(c, h.log, err, "validating a language failed")
		return
	}

	// The default is what every page is drawn in before somebody chooses, so
	// it is moved to another language rather than simply taken away.
	if wasDefault && !language.IsDefault {
		respond.Fail(c, keepADefault)
		return
	}

	if err := h.store.SaveLanguage(c.Request.Context(), language); err != nil {
		respond.Failure(c, h.log, err, "updating a language failed")
		return
	}

	h.audit.RecordWith(c, "language.updated", targetType, language.ID.String(), map[string]any{
		"language": language.Code, "enabled": language.Enabled, "is_default": language.IsDefault,
	})

	h.answer(c, http.StatusOK, language)
}

// Delete removes a language and its text. Anybody who had chosen it is shown
// the default from their next page on.
func (h *Handler) Delete(c *gin.Context) {
	language, ok := h.find(c)
	if !ok {
		return
	}

	err := h.store.DeleteLanguage(c.Request.Context(), language)
	switch {
	case errors.Is(err, store.ErrProtectedLanguage) && language.Code == model.BaseLanguage:
		respond.Fail(c, baseStays)
		return
	case errors.Is(err, store.ErrProtectedLanguage):
		respond.Fail(c, defaultStays)
		return
	case err != nil:
		respond.Failure(c, h.log, err, "deleting a language failed")
		return
	}

	h.audit.RecordWith(c, "language.deleted", targetType, language.ID.String(), map[string]any{
		"language": language.Code,
	})

	c.Status(http.StatusNoContent)
}

// Translation returns one language's text for one app, for the editor.
func (h *Handler) Translation(c *gin.Context) {
	language, app, ok := h.findWithApp(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	messages, err := h.store.Translation(ctx, language.ID, string(app))
	if err != nil {
		respond.Failure(c, h.log, err, "loading a translation failed")
		return
	}

	base, err := h.store.Language(ctx, model.BaseLanguage)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		respond.Failure(c, h.log, err, "loading the base language failed")
		return
	}

	english := i18n.Resolve(app)
	if base != nil {
		if english, err = h.store.ResolvedTranslation(ctx, base, app); err != nil {
			respond.Failure(c, h.log, err, "loading the base language failed")
			return
		}
	}

	c.JSON(http.StatusOK, translationResponse{
		App:      app,
		Keys:     i18n.Keys(app),
		Base:     english,
		Messages: messages,
	})
}

// SaveTranslation replaces one language's text for one app. Keys the app does
// not look up are dropped and counted rather than refused, so a file from an
// older release still imports.
func (h *Handler) SaveTranslation(c *gin.Context) {
	language, app, ok := h.findWithApp(c)
	if !ok {
		return
	}

	var req translationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respond.Fail(c, respond.InvalidBody)
		return
	}
	if err := validate.Struct(req); err != nil {
		respond.Failure(c, h.log, err, "validating a translation failed")
		return
	}

	messages, ignored := i18n.Known(app, req.Messages)

	if key := model.TooLongMessage(messages); key != "" {
		respond.Fail(c, translationTooLong, "key", key, "max", model.MaxMessageLength)
		return
	}

	if err := h.store.SaveTranslation(c.Request.Context(), language, string(app), messages); err != nil {
		respond.Failure(c, h.log, err, "saving a translation failed")
		return
	}

	h.audit.RecordWith(c, "language.translated", targetType, language.ID.String(), map[string]any{
		"language": language.Code, "app": string(app), "keys": len(messages), "ignored": ignored,
	})

	one, err := h.describe(c, language)
	if err != nil {
		respond.Failure(c, h.log, err, "loading translations failed")
		return
	}

	c.JSON(http.StatusOK, savedResponse{Language: one, Ignored: ignored})
}

// find loads the language named in the path, answering the request itself if
// there is no such language.
func (h *Handler) find(c *gin.Context) (*model.Language, bool) {
	language, err := h.store.Language(c.Request.Context(), c.Param("code"))

	switch {
	case errors.Is(err, store.ErrNotFound):
		respond.Fail(c, respond.LanguageNotFound)
		return nil, false
	case err != nil:
		respond.Failure(c, h.log, err, "loading a language failed")
		return nil, false
	}

	return language, true
}

// findWithApp is find, and the app named in the path as well. The sign-in
// pages are the only app a language is translated for, so any other name in
// the path is simply not an app.
func (h *Handler) findWithApp(c *gin.Context) (*model.Language, i18n.App, bool) {
	app, known := i18n.ParseApp(c.Param("app"))
	if !known {
		respond.Fail(c, noSuchApp)
		return nil, "", false
	}

	language, ok := h.find(c)

	return language, app, ok
}

// answer writes one language as the panel lists it.
func (h *Handler) answer(c *gin.Context, status int, language *model.Language) {
	one, err := h.describe(c, language)
	if err != nil {
		respond.Failure(c, h.log, err, "loading translations failed")
		return
	}

	c.JSON(status, response{Language: one})
}

// describe is one language as the panel lists it, coverage and all.
func (h *Handler) describe(c *gin.Context, language *model.Language) (languageResponse, error) {
	text, err := h.store.Translations(c.Request.Context(), language.ID)
	if err != nil {
		return languageResponse{}, err
	}

	shipped, err := i18n.Shipped()
	if err != nil {
		return languageResponse{}, err
	}

	codes := make(map[string]bool, len(shipped))
	for _, file := range shipped {
		codes[file.Code] = true
	}

	return newLanguageResponse(*language, text[language.ID], codes), nil
}

// firstOf is the value when there is one, and the fallback otherwise.
func firstOf(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}

	return fallback
}
