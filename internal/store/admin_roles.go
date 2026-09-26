package store

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"loginer/internal/model"
)

// AdminRoles returns every admin role matching a search, sorted by name.
// There are few enough of them that the panel shows all at once.
func (s *Store) AdminRoles(ctx context.Context, search string) ([]model.Role, error) {
	query := s.db.WithContext(ctx).Model(&model.Role{})

	if search = strings.TrimSpace(search); search != "" {
		like := contains(search)
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}

	var roles []model.Role
	err := query.Order("name").Find(&roles).Error

	return roles, err
}

// AdminRole returns one admin role by id.
func (s *Store) AdminRole(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role
	if err := s.db.WithContext(ctx).First(&role, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}

	return &role, nil
}

// AdminRolesByID returns the admin roles with these ids, sorted by name. An
// id named twice is one role, and an id with no role behind it is ErrNotFound.
func (s *Store) AdminRolesByID(ctx context.Context, ids []uuid.UUID) ([]model.Role, error) {
	roles := []model.Role{}
	if err := s.byID(ctx, ids, &roles, func() int { return len(roles) }); err != nil {
		return nil, err
	}

	return roles, nil
}

// AdminRoleMemberCounts says how many administrators hold each admin role,
// anywhere, by id. A role nobody holds is not in the map.
func (s *Store) AdminRoleMemberCounts(ctx context.Context) (map[uuid.UUID]int64, error) {
	var rows []struct {
		ID    uuid.UUID
		Count int64
	}

	err := s.db.WithContext(ctx).
		Table("admin_role_assignments").
		Select("role_id AS id, COUNT(DISTINCT admin_user_id) AS count").
		Group("role_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		counts[row.ID] = row.Count
	}

	return counts, nil
}

// CreateAdminRole writes a new admin role.
func (s *Store) CreateAdminRole(ctx context.Context, role *model.Role) error {
	return translate(s.db.WithContext(ctx).Create(role).Error)
}

// SaveAdminRole writes an admin role back.
func (s *Store) SaveAdminRole(ctx context.Context, role *model.Role) error {
	return translate(s.db.WithContext(ctx).Save(role).Error)
}

// DeleteAdminRole removes an admin role for good, and every assignment of it.
func (s *Store) DeleteAdminRole(ctx context.Context, role *model.Role) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM admin_role_assignments WHERE role_id = ?", role.ID).Error; err != nil {
			return err
		}

		return tx.Unscoped().Delete(role).Error
	})
}
