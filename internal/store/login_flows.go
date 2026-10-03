package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"loginer/internal/cache"
	"loginer/internal/model"
)

// LoginFlows returns every flow, the default first and the rest by name, which
// is the order the panel lists them in.
func (s *Store) LoginFlows(ctx context.Context) ([]model.LoginFlow, error) {
	var flows []model.LoginFlow
	err := s.db.WithContext(ctx).Order("is_default DESC, name").Find(&flows).Error

	return flows, err
}

// LoginFlow returns one flow by id.
func (s *Store) LoginFlow(ctx context.Context, id uuid.UUID) (*model.LoginFlow, error) {
	return cached(ctx, s, cache.LoginFlows, "id:"+id.String(), func() (*model.LoginFlow, error) {
		var flow model.LoginFlow
		if err := s.db.WithContext(ctx).First(&flow, "id = ?", id).Error; err != nil {
			return nil, translate(err)
		}

		return &flow, nil
	})
}

// DefaultLoginFlow returns the flow used by applications without their own,
// writing one if none exists.
func (s *Store) DefaultLoginFlow(ctx context.Context) (*model.LoginFlow, error) {
	var flow model.LoginFlow
	if s.cache.Get(ctx, cache.LoginFlows, "default", &flow) {
		return &flow, nil
	}

	err := translate(s.db.WithContext(ctx).Order("created_at").First(&flow, "is_default = ?", true).Error)
	switch {
	case err == nil:
		s.cache.Set(ctx, cache.LoginFlows, "default", flow)
		return &flow, nil
	case !errors.Is(err, ErrNotFound):
		return nil, err
	}

	flow = model.DefaultLoginFlow()
	if err := s.db.WithContext(ctx).Create(&flow).Error; err != nil {
		return nil, translate(err)
	}

	s.forget(ctx, cache.LoginFlows)

	return &flow, nil
}

// EffectiveLoginFlow returns the application's flow if it is still enabled,
// otherwise the default.
func (s *Store) EffectiveLoginFlow(ctx context.Context, app *model.Application) (*model.LoginFlow, error) {
	if app != nil && app.LoginFlowID != nil {
		flow, err := s.LoginFlow(ctx, *app.LoginFlowID)
		switch {
		case err == nil && flow.IsEnabled:
			return flow, nil
		case err != nil && !errors.Is(err, ErrNotFound):
			return nil, err
		}
	}

	return s.DefaultLoginFlow(ctx)
}

// CreateLoginFlow adds a flow.
func (s *Store) CreateLoginFlow(ctx context.Context, flow *model.LoginFlow) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(flow).Error; err != nil {
			return translate(err)
		}

		return demoteOtherDefaults(tx, flow)
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.LoginFlows)

	return nil
}

// SaveLoginFlow writes a flow back.
func (s *Store) SaveLoginFlow(ctx context.Context, flow *model.LoginFlow) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(flow).Error; err != nil {
			return translate(err)
		}

		return demoteOtherDefaults(tx, flow)
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.LoginFlows)

	return nil
}

// demoteOtherDefaults keeps exactly one default flow: the one just written
// wins.
func demoteOtherDefaults(tx *gorm.DB, flow *model.LoginFlow) error {
	if !flow.IsDefault {
		return nil
	}

	err := tx.Model(&model.LoginFlow{}).
		Where("id <> ? AND is_default = ?", flow.ID, true).
		Update("is_default", false).Error

	return translate(err)
}

// ErrDefaultLoginFlow protects the default flow from removal.
var ErrDefaultLoginFlow = errors.New("the default flow cannot be removed")

// DeleteLoginFlow removes a flow, and lets the applications that were using
// it fall back to the default.
func (s *Store) DeleteLoginFlow(ctx context.Context, flow *model.LoginFlow) error {
	if flow.IsDefault {
		return ErrDefaultLoginFlow
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Application{}).
			Where("login_flow_id = ?", flow.ID).
			Update("login_flow_id", nil).Error
		if err != nil {
			return translate(err)
		}

		return translate(tx.Unscoped().Delete(flow).Error)
	})
	if err != nil {
		return err
	}

	s.forget(ctx, cache.LoginFlows, cache.Clients)

	return nil
}

// LoginFlowApplications counts the applications pointed at each flow, by flow
// id. A flow nothing uses is left out.
func (s *Store) LoginFlowApplications(ctx context.Context) (map[uuid.UUID]int, error) {
	var rows []struct {
		LoginFlowID uuid.UUID
		Count       int
	}

	err := s.db.WithContext(ctx).
		Model(&model.Application{}).
		Select("login_flow_id, count(*) as count").
		Where("login_flow_id IS NOT NULL").
		Group("login_flow_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		counts[row.LoginFlowID] = row.Count
	}

	return counts, nil
}
