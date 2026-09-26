package store

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"loginer/internal/model"
)

// RoleQuery is what a listing of roles asks for: which scopes, a search box,
// a filter, and a page.
type RoleQuery struct {
	// Global includes the global roles.
	Global bool
	// Applications includes these applications' roles: nil means every
	// application's, and an empty list none.
	Applications []uuid.UUID
	// Search matches the name or the description.
	Search string
	// Default filters on the is_default flag. Nil means both.
	Default *bool
	Limit   int
	Offset  int
}

// byName orders what a preload brings back, so the roles a user holds or a
// role inherits always come out in the same order.
func byName(db *gorm.DB) *gorm.DB {
	return db.Order("name")
}

// UserRoles returns a page of roles, sorted by name, along with how many
// match the query in total.
func (s *Store) UserRoles(ctx context.Context, q RoleQuery) ([]model.UserRole, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.UserRole{})

	scopes := s.db.Where("1 = 0")
	if q.Global {
		scopes = scopes.Or("application_id IS NULL")
	}
	switch {
	case q.Applications == nil:
		scopes = scopes.Or("application_id IS NOT NULL")
	case len(q.Applications) > 0:
		scopes = scopes.Or("application_id IN ?", q.Applications)
	}
	query = query.Where(scopes)

	if search := strings.TrimSpace(q.Search); search != "" {
		like := contains(search)
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}

	if q.Default != nil {
		query = query.Where("is_default = ?", *q.Default)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Global roles first, then each application's roles together, by the
	// application's name, so a list that mixes applications reads as groups.
	var roles []model.UserRole
	err := query.
		Preload("Inherits", byName).
		Preload("APIScopes", scopesByName).
		Order("application_id IS NOT NULL").
		Order("(SELECT LOWER(applications.name) FROM applications WHERE applications.id = user_roles.application_id)").
		Order("name").
		Limit(q.Limit).
		Offset(q.Offset).
		Find(&roles).Error

	return roles, total, err
}

// UserRole returns one role by id, with the roles it inherits.
func (s *Store) UserRole(ctx context.Context, id uuid.UUID) (*model.UserRole, error) {
	var role model.UserRole
	err := s.db.WithContext(ctx).
		Preload("Inherits", byName).
		Preload("APIScopes", scopesByName).
		First(&role, "id = ?", id).Error
	if err != nil {
		return nil, translate(err)
	}

	return &role, nil
}

// UserRolesByID returns the roles with these ids, sorted by name. An id named
// twice is one role, and an id with no role behind it is ErrNotFound.
func (s *Store) UserRolesByID(ctx context.Context, ids []uuid.UUID) ([]model.UserRole, error) {
	roles := []model.UserRole{}

	if err := s.byID(ctx, ids, &roles, func() int { return len(roles) }); err != nil {
		return nil, err
	}

	return roles, nil
}

// DefaultUserRoles returns the roles every new user is given: the default
// global roles, and each enabled application's default roles.
func (s *Store) DefaultUserRoles(ctx context.Context) ([]model.UserRole, error) {
	var roles []model.UserRole
	err := s.db.WithContext(ctx).
		Where("is_default = ?", true).
		Where("application_id IS NULL OR application_id IN (SELECT id FROM applications WHERE enabled)").
		Order("application_id NULLS FIRST, name").
		Find(&roles).Error

	return roles, err
}

// AddUserRoles gives a user roles, directly. A role the user already holds is
// left as it is.
func (s *Store) AddUserRoles(ctx context.Context, userID uuid.UUID, roles []model.UserRole) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, role := range roles {
			err := tx.Exec(
				"INSERT INTO user_role_members (user_id, user_role_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
				userID, role.ID,
			).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// RemoveUserRole takes away a role the user was given directly. The roles it
// included stop coming to the user through it; any that another held role
// still includes stay.
func (s *Store) RemoveUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.db.WithContext(ctx).
		Exec("DELETE FROM user_role_members WHERE user_id = ? AND user_role_id = ?", userID, roleID).
		Error
}

// RoleGraph loads every role with the roles it inherits, which is what
// resolving inheritance and refusing a cycle need.
func (s *Store) RoleGraph(ctx context.Context) (model.RoleGraph, error) {
	var roles []model.UserRole
	err := s.db.WithContext(ctx).
		Preload("Inherits", byName).
		Preload("APIScopes", scopesByName).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}

	graph := make(model.RoleGraph, len(roles))
	for _, role := range roles {
		graph[role.ID] = role
	}

	return graph, nil
}

// RoleMemberCounts says how many users hold each of these roles directly, by
// id. A role nobody holds is not in the map.
func (s *Store) RoleMemberCounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]int64{}, nil
	}

	return s.countBy(ctx, "user_role_members", "user_role_id", ids)
}

// CreateUserRole writes a new role with its inheritance. The inherited roles
// already exist, so only the joins are written.
func (s *Store) CreateUserRole(ctx context.Context, role *model.UserRole) error {
	return translate(s.db.WithContext(ctx).Omit("Inherits.*", "APIScopes.*").Create(role).Error)
}

// SaveUserRole writes a role back, replacing what it inherits and the API
// scopes it grants with the ones it carries. It is one transaction, so a role
// is never left with half of an edit.
func (s *Store) SaveUserRole(ctx context.Context, role *model.UserRole) error {
	return translate(s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Save(role).Error; err != nil {
			return err
		}

		if err := tx.Model(role).Association("Inherits").Replace(role.Inherits); err != nil {
			return err
		}

		return tx.Model(role).Association("APIScopes").Replace(role.APIScopes)
	}))
}

// DeleteUserRole removes a role for good, along with every inheritance to or
// from it, every user's hold on it, and the API scopes it grants. The joins
// are removed by hand rather than left to the foreign keys, so this does not
// depend on how those constraints were created.
func (s *Store) DeleteUserRole(ctx context.Context, role *model.UserRole) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"DELETE FROM user_role_inherits WHERE role_id = @id OR inherited_role_id = @id",
			"DELETE FROM user_role_members WHERE user_role_id = @id",
			"DELETE FROM user_role_api_scopes WHERE user_role_id = @id",
		}

		for _, statement := range statements {
			if err := tx.Exec(statement, map[string]any{"id": role.ID}).Error; err != nil {
				return err
			}
		}

		return tx.Unscoped().Delete(role).Error
	})
}

// byID loads the rows with these ids into `into`, sorted by name, and says
// ErrNotFound if any of them is missing. `found` reports how many were loaded.
func (s *Store) byID(ctx context.Context, ids []uuid.UUID, into any, found func() int) error {
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		seen[id] = true
	}

	if len(seen) == 0 {
		return nil
	}

	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Order("name").Find(into).Error; err != nil {
		return err
	}

	if found() != len(seen) {
		return ErrNotFound
	}

	return nil
}

// countBy counts the rows of a join table per value of one of its columns,
// optionally only for some values.
func (s *Store) countBy(ctx context.Context, table, column string, only []uuid.UUID) (map[uuid.UUID]int64, error) {
	var rows []struct {
		ID    uuid.UUID
		Count int64
	}

	query := s.db.WithContext(ctx).Table(table).Select(column + " AS id, COUNT(*) AS count").Group(column)
	if only != nil {
		query = query.Where(column+" IN ?", only)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int64, len(rows))
	for _, row := range rows {
		counts[row.ID] = row.Count
	}

	return counts, nil
}
