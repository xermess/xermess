package otp

import (
	"xermess/internal/model"
)

// settingsResponse is the settings as the panel sees them, built by hand so
// the row's own bookkeeping stays out of a record of settings.
type settingsResponse struct {
	Length          int `json:"length"`
	LifetimeMinutes int `json:"lifetime_minutes"`
	MaxAttempts     int `json:"max_attempts"`
	ResendSeconds   int `json:"resend_seconds"`
}

// usingFlow is one login flow that asks for an emailed code, so the page can
// say who these settings are for rather than leaving an administrator to
// guess whether anything reads them.
type usingFlow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Enabled   bool   `json:"enabled"`
	IsDefault bool   `json:"is_default"`
}

// limits are the bounds the panel holds its fields to, from the model, so the
// two cannot disagree about what a code may be.
type limits struct {
	MinLength       int `json:"min_length"`
	MaxLength       int `json:"max_length"`
	MaxLifetime     int `json:"max_lifetime_minutes"`
	MaxAttempts     int `json:"max_attempts"`
	MaxResendSecond int `json:"max_resend_seconds"`
}

// response is what both endpoints answer with.
type response struct {
	OTP    settingsResponse `json:"otp"`
	Limits limits           `json:"limits"`
	Flows  []usingFlow      `json:"flows"`
}

func newResponse(settings model.OTPSettings, flows []model.LoginFlow) response {
	using := []usingFlow{}
	for _, flow := range flows {
		if !flow.Offers(model.StepEmailCode) {
			continue
		}

		using = append(using, usingFlow{
			ID:        flow.ID.String(),
			Name:      flow.Name,
			Slug:      flow.Slug,
			Enabled:   flow.Enabled,
			IsDefault: flow.IsDefault,
		})
	}

	return response{
		OTP: settingsResponse{
			Length:          settings.Length,
			LifetimeMinutes: settings.LifetimeMinutes,
			MaxAttempts:     settings.MaxAttempts,
			ResendSeconds:   settings.ResendSeconds,
		},
		Limits: limits{
			MinLength:       model.MinOTPLength,
			MaxLength:       model.MaxOTPLength,
			MaxLifetime:     model.MaxOTPLifetimeMinutes,
			MaxAttempts:     model.MaxOTPAttempts,
			MaxResendSecond: model.MaxOTPResendSeconds,
		},
		Flows: using,
	}
}
