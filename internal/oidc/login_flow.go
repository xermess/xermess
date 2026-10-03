package oidc

import (
	"context"

	"loginer/internal/model"
)

// LoginOptions is what the sign-in pages may offer: registration, password
// reset, and the flow's steps. The flow is the application's if the handle
// names one, otherwise the default, including for an expired handle.
func (s *Service) LoginOptions(ctx context.Context, handle string) (model.PublicLoginOptions, error) {
	flow, err := s.flowFor(ctx, handle)
	if err != nil {
		return model.PublicLoginOptions{}, err
	}

	return flow.PublicOptions(), nil
}

// flowFor is the flow a sign-in follows: its application's, or the default.
// Every way in asks, so the rules hold however someone arrives.
func (s *Service) flowFor(ctx context.Context, handle string) (*model.LoginFlow, error) {
	var app *model.Application

	if handle != "" {
		if req, err := s.pending(ctx, handle); err == nil {
			app = req.Application
		}
	}

	return s.store.EffectiveLoginFlow(ctx, app)
}
