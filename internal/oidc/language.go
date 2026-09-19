package oidc

import (
	"context"

	"xermess/internal/model"
	"xermess/internal/store"
	"xermess/locales"
)

// PublicLanguage is one language the sign-in pages may be shown in, as the
// picker lists it: the code the pages ask for its text by, and the two names
// — the language in English, and in itself.
type PublicLanguage struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Native string `json:"native"`
}

func publicLanguage(language model.Language) PublicLanguage {
	return PublicLanguage{Code: language.Code, Name: language.Name, Native: language.Native}
}

// Languages is what the sign-in pages ask for when they draw their language
// picker: the languages this installation offers, in the order it offers
// them, and which of them somebody gets before they have chosen.
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

// textIn is the sign-in pages' text in a language, for what the server writes
// itself — an email. A language that is not offered, or none given, is the
// default one; and when even that cannot be read, the text is the shipped
// English, so an email always goes out in something.
func (s *Service) textIn(ctx context.Context, code string) map[string]string {
	language, err := s.store.Language(ctx, code)
	if err != nil || !language.Enabled {
		language, err = s.store.DefaultLanguage(ctx)
	}

	if err == nil {
		if text, err := s.store.ResolvedTranslation(ctx, language, locales.ID); err == nil {
			return text
		}
	}

	return locales.Resolve(locales.ID)
}

// LanguageText is the text the sign-in pages are drawn with in one language:
// every key they look up, a missing translation already filled in from the
// base language, so the pages have nothing to work out.
//
// Only a language that is offered: one that is off is somebody's work in
// progress, and store.ErrNotFound is what anything else gets.
func (s *Service) LanguageText(ctx context.Context, code string) (PublicLanguage, map[string]string, error) {
	language, err := s.store.Language(ctx, code)
	if err != nil {
		return PublicLanguage{}, nil, err
	}

	if !language.Enabled {
		return PublicLanguage{}, nil, store.ErrNotFound
	}

	messages, err := s.store.ResolvedTranslation(ctx, language, locales.ID)
	if err != nil {
		return PublicLanguage{}, nil, err
	}

	return publicLanguage(*language), messages, nil
}
