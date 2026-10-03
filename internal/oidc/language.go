package oidc

import (
	"context"

	"loginer/i18n"
	"loginer/internal/model"
	"loginer/internal/store"
)

// PublicLanguage is one language in the picker: its code and its English and
// native names.
type PublicLanguage struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
}

func publicLanguage(language model.Language) PublicLanguage {
	return PublicLanguage{Code: language.Code, Name: language.Name, NativeName: language.NativeName}
}

// Languages returns the offered languages in order, and the default.
func (s *Service) Languages(ctx context.Context) ([]PublicLanguage, string, error) {
	offered, err := s.store.OfferedLanguages(ctx)
	if err != nil {
		return nil, "", err
	}

	fallback, err := s.store.DefaultLanguage(ctx)
	if err != nil {
		return nil, "", err
	}

	languages := make([]PublicLanguage, 0, len(offered))
	for _, language := range offered {
		languages = append(languages, publicLanguage(language))
	}

	return languages, fallback.Code, nil
}

// textIn is the pages' text in a language, for emails. It falls back to the
// default language, then shipped English.
func (s *Service) textIn(ctx context.Context, code string) map[string]string {
	language, err := s.store.Language(ctx, code)
	if err != nil || !language.IsEnabled {
		language, err = s.store.DefaultLanguage(ctx)
	}

	if err == nil {
		if text, err := s.store.ResolvedTranslation(ctx, language, i18n.ID); err == nil {
			return text
		}
	}

	return i18n.Resolve(i18n.ID)
}

// LanguageText is every key the sign-in pages use in one offered language, with
// gaps filled from the base language.
func (s *Service) LanguageText(ctx context.Context, code string) (PublicLanguage, map[string]string, error) {
	language, err := s.store.Language(ctx, code)
	if err != nil {
		return PublicLanguage{}, nil, err
	}

	if !language.IsEnabled {
		return PublicLanguage{}, nil, store.ErrNotFound
	}

	messages, err := s.store.ResolvedTranslation(ctx, language, i18n.ID)
	if err != nil {
		return PublicLanguage{}, nil, err
	}

	return publicLanguage(*language), messages, nil
}
