package store

import (
	"context"
	"errors"

	"loginer/internal/cache"
	"loginer/internal/model"
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

	loaded, err := s.OrganizationForUpdate(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cache.Organization, "settings", *loaded)

	return loaded, nil
}

// OrganizationForUpdate reads the organisation from the database, never from
// the cache: what is about to be changed and written back has to be the row
// as it is, or a stale copy — one cached before the database was reset, or
// by another server sharing the Redis — is saved over it.
func (s *Store) OrganizationForUpdate(ctx context.Context) (*model.Organization, error) {
	var organization model.Organization

	err := translate(s.db.WithContext(ctx).Order("created_at").First(&organization).Error)
	switch {
	case err == nil:
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
//
// It updates the row it was read from and nothing else. GORM's Save would
// insert a second organisation when that row is not there; this says so
// instead, as ErrNotFound.
func (s *Store) SaveOrganization(ctx context.Context, organization *model.Organization) error {
	result := s.db.WithContext(ctx).Model(organization).Select("*").Omit("created_at").Updates(organization)
	if err := translate(result.Error); err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	s.forget(ctx, cache.Organization)

	return nil
}
