package store

import (
	"context"
	"errors"

	"xermess/internal/cache"
	"xermess/internal/model"
)

// Organization returns the organisation this installation belongs to.
//
// There is one, so it is read by nothing but its age: the oldest row is the
// one the migration seeded. A database that somehow holds none gets the
// default written for it, which is what keeps the settings page working on an
// installation whose row was removed by hand rather than answering 404 for
// something that cannot be created from the panel.
//
// Every sign-in page asks for it, so it is read through the cache.
func (s *Store) Organization(ctx context.Context) (*model.Organization, error) {
	var organization model.Organization
	if s.cache.Get(ctx, cache.Organization, "settings", &organization) {
		return &organization, nil
	}

	err := translate(s.db.WithContext(ctx).Order("created_at").First(&organization).Error)
	switch {
	case err == nil:
		s.cache.Set(ctx, cache.Organization, "settings", organization)
		return &organization, nil
	case !errors.Is(err, ErrNotFound):
		return nil, err
	}

	organization = model.DefaultOrganization()
	if err := s.db.WithContext(ctx).Create(&organization).Error; err != nil {
		return nil, translate(err)
	}

	s.forget(ctx, cache.Organization)

	return &organization, nil
}

// SaveOrganization writes the organisation back.
func (s *Store) SaveOrganization(ctx context.Context, organization *model.Organization) error {
	if err := translate(s.db.WithContext(ctx).Save(organization).Error); err != nil {
		return err
	}

	s.forget(ctx, cache.Organization)

	return nil
}
