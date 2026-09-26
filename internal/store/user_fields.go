package store

import (
	"context"

	"github.com/google/uuid"

	"loginer/internal/model"
)

// UserFields returns every field a user record has, in the order they are
// shown. The panel builds both its table and its form from this.
func (s *Store) UserFields(ctx context.Context) ([]model.UserField, error) {
	var fields []model.UserField
	err := s.db.WithContext(ctx).Order("position, created_at").Find(&fields).Error

	return fields, err
}

// UserField returns one field by id.
func (s *Store) UserField(ctx context.Context, id uuid.UUID) (*model.UserField, error) {
	var field model.UserField
	if err := s.db.WithContext(ctx).First(&field, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}

	return &field, nil
}

// CreateUserField adds a field to every user record.
func (s *Store) CreateUserField(ctx context.Context, field *model.UserField) error {
	return translate(s.db.WithContext(ctx).Create(field).Error)
}

// SaveUserField writes a field back.
func (s *Store) SaveUserField(ctx context.Context, field *model.UserField) error {
	return translate(s.db.WithContext(ctx).Save(field).Error)
}

// DeleteUserField removes a field. The values stored under its name stay in
// the user records until those are next saved, at which point they are
// dropped: removing a column from the panel destroys nothing by itself.
func (s *Store) DeleteUserField(ctx context.Context, field *model.UserField) error {
	return s.db.WithContext(ctx).Unscoped().Delete(field).Error
}

// NextFieldPosition is where a new field goes: after the last one.
func (s *Store) NextFieldPosition(ctx context.Context) int {
	var last model.UserField
	if err := s.db.WithContext(ctx).Order("position DESC").First(&last).Error; err != nil {
		return 1
	}

	return last.Position + 1
}
