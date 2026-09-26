package oidc

import (
	"context"

	"loginer/internal/model"
)

// LoginOptions is what the sign-in pages are told they may offer: whether an
// account can be made, whether a password can be reset, and which steps the
// flow names.
//
// Which flow that is depends on who is signing somebody in. A sign-in under
// way names an application, and an application may name a flow of its own; a
// page opened on its own — a reset link, a bare sign-in — gets the default,
// and so does a handle that has expired. A sign-in page that cannot tell
// which flow applies should show the installation's usual one rather than
// nothing, which is why an unusable handle is not an error here.
func (s *Service) LoginOptions(ctx context.Context, handle string) (model.PublicLoginOptions, error) {
	flow, err := s.flowFor(ctx, handle)
	if err != nil {
		return model.PublicLoginOptions{}, err
	}

	return flow.PublicOptions(), nil
}

// flowFor is the login flow a sign-in goes through: its application's, or
// the default for a sign-in with no application waiting or a handle that no
// longer works. Every way in asks — a password, a provider, an organisation's
// identity provider — so a flow's rules hold however somebody arrives.
func (s *Service) flowFor(ctx context.Context, handle string) (*model.LoginFlow, error) {
	var app *model.Application

	if handle != "" {
		if req, err := s.pending(ctx, handle); err == nil {
			app = req.Application
		}
	}

	return s.store.EffectiveLoginFlow(ctx, app)
}
