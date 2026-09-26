package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"loginer/i18n"
	"loginer/internal/cache"
	"loginer/internal/model"
)

// ErrProtectedLanguage is returned for removing the base language, or the
// default one: the first is what every missing translation falls back to, and
// the second is what somebody sees before they have chosen.
var ErrProtectedLanguage = errors.New("this language cannot be removed")

// Languages returns every language, in the order the picker lists them.
func (s *Store) Languages(ctx context.Context) ([]model.Language, error) {
	return cached(ctx, s, cache.Languages, "all", func() ([]model.Language, error) {
		var languages []model.Language
		err := s.db.WithContext(ctx).Order("position, code").Find(&languages).Error

		return languages, err
	})
}

// Language returns one language by code.
func (s *Store) Language(ctx context.Context, code string) (*model.Language, error) {
	return cached(ctx, s, cache.Languages, "code:"+code, func() (*model.Language, error) {
		var language model.Language
		if err := s.db.WithContext(ctx).First(&language, "code = ?", code).Error; err != nil {
			return nil, translate(err)
		}

		return &language, nil
	})
}

// DefaultLanguage returns the language somebody sees before they have chosen
// one.
//
// A database holding none gets the base language written for it, as the
// organisation does: every sign-in page asks for this, and an installation
// whose row was removed by hand should still draw itself in something.
func (s *Store) DefaultLanguage(ctx context.Context) (*model.Language, error) {
	var found model.Language
	if s.cache.Get(ctx, cache.Languages, "default", &found) {
		return &found, nil
	}

	var language model.Language

	err := translate(s.db.WithContext(ctx).Order("position, code").First(&language, "is_default = ?", true).Error)
	switch {
	case err == nil:
		s.cache.Set(ctx, cache.Languages, "default", language)
		return &language, nil
	case !errors.Is(err, ErrNotFound):
		return nil, err
	}

	language = model.DefaultLanguage()
	if err := s.db.WithContext(ctx).Create(&language).Error; err != nil {
		return nil, translate(err)
	}

	s.forget(ctx, cache.Languages)

	return &language, nil
}

// OfferedLanguages returns the languages the sign-in pages may offer, in
// order: the enabled ones, which always includes the default.
func (s *Store) OfferedLanguages(ctx context.Context) ([]model.Language, error) {
	if _, err := s.DefaultLanguage(ctx); err != nil {
		return nil, err
	}

	return cached(ctx, s, cache.Languages, "offered", func() ([]model.Language, error) {
		var languages []model.Language
		err := s.db.WithContext(ctx).Where("enabled = ?", true).Order("position, code").Find(&languages).Error

		return languages, err
	})
}

// CreateLanguage writes a new language and whatever text it starts with, by
// app. Nothing is written unless all of it is.
func (s *Store) CreateLanguage(ctx context.Context, language *model.Language, messages map[string]map[string]string) error {
	language.Code = strings.TrimSpace(language.Code)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(language).Error; err != nil {
			return translate(err)
		}

		for app, text := range messages {
			if err := saveTranslation(tx, language.ID, app, text); err != nil {
				return err
			}
		}

		return demoteOtherLanguages(tx, language)
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.Languages)

	return nil
}

// SaveLanguage writes a language's settings back.
func (s *Store) SaveLanguage(ctx context.Context, language *model.Language) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(language).Error; err != nil {
			return translate(err)
		}

		return demoteOtherLanguages(tx, language)
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.Languages)

	return nil
}

// DeleteLanguage removes a language and all of its text. The base language
// and the default are refused: see ErrProtectedLanguage.
func (s *Store) DeleteLanguage(ctx context.Context, language *model.Language) error {
	if language.Code == model.BaseLanguage || language.IsDefault {
		return ErrProtectedLanguage
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unscoped, like every delete here: the code is unique, and a
		// soft-deleted row would keep the code taken from a language added
		// again under it.
		if err := tx.Unscoped().Where("language_id = ?", language.ID).Delete(&model.Translation{}).Error; err != nil {
			return translate(err)
		}

		return translate(tx.Unscoped().Delete(language).Error)
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.Languages)

	return nil
}

// demoteOtherLanguages keeps exactly one language marked as the default:
// whichever was just written wins, and any other holding the mark loses it.
//
// Making the newest one win is what lets the panel offer a plain switch, as
// it does for login flows. The alternative — refusing a second default —
// would mean unmarking the old one first, for a rule the server can keep.
func demoteOtherLanguages(tx *gorm.DB, language *model.Language) error {
	if !language.IsDefault {
		return nil
	}

	err := tx.Model(&model.Language{}).
		Where("id <> ? AND is_default = ?", language.ID, true).
		Update("is_default", false).Error

	return translate(err)
}

// NextLanguagePosition is where a new language goes in the picker: after the
// last.
func (s *Store) NextLanguagePosition(ctx context.Context) int {
	return nextPosition(s.db.WithContext(ctx))
}

// Translations returns the text of the given languages, by language id and
// then by app. A language or an app with no text has no entry.
func (s *Store) Translations(ctx context.Context, languages ...uuid.UUID) (map[uuid.UUID]map[string]map[string]string, error) {
	out := map[uuid.UUID]map[string]map[string]string{}
	if len(languages) == 0 {
		return out, nil
	}

	var rows []model.Translation
	if err := s.db.WithContext(ctx).Where("language_id IN ?", languages).Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		if out[row.LanguageID] == nil {
			out[row.LanguageID] = map[string]map[string]string{}
		}
		out[row.LanguageID][row.App] = row.Messages
	}

	return out, nil
}

// Translation returns one language's text for one app, which is empty rather
// than missing when nothing has been translated.
func (s *Store) Translation(ctx context.Context, language uuid.UUID, app string) (map[string]string, error) {
	var row model.Translation

	err := translate(s.db.WithContext(ctx).First(&row, "language_id = ? AND app = ?", language, app).Error)
	switch {
	case errors.Is(err, ErrNotFound):
		return map[string]string{}, nil
	case err != nil:
		return nil, err
	}

	return row.Messages, nil
}

// ResolvedTranslation is what an app is sent for one language: every key it
// looks up, from the language's own text, then the base language's as the
// database holds it, then the shipped base file (i18n.Resolve).
//
// It is the largest thing the pages ask for — every key of an app, on every
// render — and the one worth caching most: it is kept whole, per language and
// app, until any language's text changes.
func (s *Store) ResolvedTranslation(ctx context.Context, language *model.Language, app i18n.App) (map[string]string, error) {
	return cached(ctx, s, cache.Languages, "text:"+language.Code+":"+string(app), func() (map[string]string, error) {
		return s.resolveTranslation(ctx, language, app)
	})
}

// resolveTranslation is ResolvedTranslation without the cache.
func (s *Store) resolveTranslation(ctx context.Context, language *model.Language, app i18n.App) (map[string]string, error) {
	own, err := s.Translation(ctx, language.ID, string(app))
	if err != nil {
		return nil, err
	}

	if language.Code == model.BaseLanguage {
		return i18n.Resolve(app, own), nil
	}

	base, err := s.Language(ctx, model.BaseLanguage)
	switch {
	case errors.Is(err, ErrNotFound):
		return i18n.Resolve(app, own), nil
	case err != nil:
		return nil, err
	}

	english, err := s.Translation(ctx, base.ID, string(app))
	if err != nil {
		return nil, err
	}

	return i18n.Resolve(app, own, english), nil
}

// SaveTranslation replaces one language's text for one app.
//
// It also moves the language's updated_at, which is the version the apps key
// their copy of the text by: a save is seen on the next page anybody opens.
func (s *Store) SaveTranslation(ctx context.Context, language *model.Language, app string, messages map[string]string) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := saveTranslation(tx, language.ID, app, messages); err != nil {
			return err
		}

		return translate(tx.Model(language).Update("updated_at", time.Now()).Error)
	})
	if err != nil {
		return err
	}

	// Every language's text, not only this one's: English is what the rest
	// fall back to, so its keys are inside every other cached text.
	s.forget(ctx, cache.Languages)

	return nil
}

// SaveTranslationKeys changes some of one language's text for one app and
// leaves the rest as it was.
//
// SaveTranslation is a whole file: the Languages page holds every key, so
// what it sends is what the language says. A page that edits a handful of
// keys — the Mail page and its email.* ones — has the rest nowhere, and
// sending what it holds would clear them. This merges instead: a key with an
// empty value is removed rather than stored blank, so clearing an override in
// the panel puts the shipped text back.
func (s *Store) SaveTranslationKeys(ctx context.Context, language *model.Language, app string, messages map[string]string) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.Translation

		err := translate(tx.First(&row, "language_id = ? AND app = ?", language.ID, app).Error)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}

		merged := map[string]string{}
		for key, text := range row.Messages {
			merged[key] = text
		}
		for key, text := range messages {
			if text == "" {
				delete(merged, key)
				continue
			}
			merged[key] = text
		}

		if err := saveTranslation(tx, language.ID, app, merged); err != nil {
			return err
		}

		return translate(tx.Model(language).Update("updated_at", time.Now()).Error)
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.Languages)

	return nil
}

// saveTranslation writes the row for one language and app, whether or not
// there is one already.
func saveTranslation(tx *gorm.DB, language uuid.UUID, app string, messages map[string]string) error {
	if messages == nil {
		messages = map[string]string{}
	}

	row := model.Translation{LanguageID: language, App: app, Messages: messages}

	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "language_id"}, {Name: "app"}},
		DoUpdates: clause.AssignmentColumns([]string{"messages", "updated_at"}),
	}).Create(&row).Error

	return translate(err)
}

// EnsureLanguages brings the database's languages up to date with the ones
// the server ships with.
//
// On the first start — no text in the database at all — every shipped
// language is imported: a row for each, off unless it is the base language,
// and all of its text. After that the database is the panel's, and a start
// only adds what a release brought: a key a shipped group has that the
// database's copy of that language does not. It never changes a message that
// is there, and never brings back a language somebody removed.
//
// The one thing that follows from that: a message cleared in the panel, in a
// language that ships, is a key the database no longer has — so the next
// start puts the shipped text back. A shipped language can be reworded; the
// way to empty it is to remove it.
func (s *Store) EnsureLanguages(ctx context.Context, shipped []i18n.File) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := dropPanelText(tx); err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&model.Translation{}).Count(&count).Error; err != nil {
			return err
		}

		first := count == 0

		for _, file := range shipped {
			var language model.Language

			err := translate(tx.First(&language, "code = ?", file.Code).Error)
			switch {
			case errors.Is(err, ErrNotFound) && !first:
				// Removed on purpose, or never imported because it shipped
				// later than this installation started. Either way it is the
				// panel's to add.
				continue
			case errors.Is(err, ErrNotFound):
				language = model.Language{
					Code: file.Code, Name: file.Name, Native: file.Native,
					Enabled:  file.Code == model.BaseLanguage,
					Position: nextPosition(tx),
				}
				if err := tx.Create(&language).Error; err != nil {
					return translate(err)
				}
			case err != nil:
				return err
			}

			// A row written before languages had names — the migration's
			// seed, on an older build — takes them from the file.
			if language.Name == "" || language.Native == "" {
				err := tx.Model(&language).Updates(map[string]any{"name": file.Name, "native": file.Native}).Error
				if err != nil {
					return translate(err)
				}
			}

			if err := addShippedKeys(tx, &language, file); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.Languages)

	return nil
}

// dropPanelText removes the panel's text from every language.
//
// The panel is written in English, in its own markup, and is not translated:
// nothing reads these rows, and an installation that started while it was
// still translated has them. Every start clears them, so one that is stepped
// forward comes out the same as a fresh one.
func dropPanelText(tx *gorm.DB) error {
	err := tx.Unscoped().
		Where("app = ?", string(i18n.Console)).
		Delete(&model.Translation{}).Error

	return translate(err)
}

// addShippedKeys copies into the database the keys a shipped catalog has and
// the database's copy of the same language does not.
func addShippedKeys(tx *gorm.DB, language *model.Language, file i18n.File) error {
	for app, shipped := range file.Messages {
		if !i18n.ServesApp(language.Code, app) {
			continue
		}

		var row model.Translation

		err := translate(tx.First(&row, "language_id = ? AND app = ?", language.ID, string(app)).Error)
		switch {
		case errors.Is(err, ErrNotFound):
			row.Messages = map[string]string{}
		case err != nil:
			return err
		}

		added := false
		for key, value := range shipped {
			if _, has := row.Messages[key]; !has && strings.TrimSpace(value) != "" {
				row.Messages[key] = value
				added = true
			}
		}

		if !added {
			continue
		}

		if err := saveTranslation(tx, language.ID, string(app), row.Messages); err != nil {
			return err
		}
	}

	return nil
}

// nextPosition is one after the last language's position, on whichever
// connection or transaction it is given.
func nextPosition(tx *gorm.DB) int {
	var last model.Language
	if err := tx.Order("position DESC").First(&last).Error; err != nil {
		return 1
	}

	return last.Position + 1
}
