package store

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"loginer/internal/model"
)

// ApplicationQuery is what a listing of applications asks for.
type ApplicationQuery struct {
	// Search matches the name, the description or the client id.
	Search string
	// Type filters on the application type. Empty means every type.
	Type model.ApplicationType
	// Enabled filters on the enabled flag. Nil means both.
	Enabled *bool
	// Only keeps these applications, for an administrator whose roles reach
	// just some of them. Nil means every application.
	Only   []uuid.UUID
	Limit  int
	Offset int
}

// Applications returns a page of applications, sorted by name, and how many
// match the query in total.
func (s *Store) Applications(ctx context.Context, q ApplicationQuery) ([]model.Application, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.Application{})

	if search := strings.TrimSpace(q.Search); search != "" {
		like := contains(search)
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(client_id) LIKE ?",
			like, like, like,
		)
	}

	if q.Type != "" {
		query = query.Where("type = ?", q.Type)
	}

	if q.Enabled != nil {
		query = query.Where("enabled = ?", *q.Enabled)
	}

	if q.Only != nil {
		if len(q.Only) == 0 {
			return []model.Application{}, 0, nil
		}
		query = query.Where("id IN ?", q.Only)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var apps []model.Application
	err := query.Order("LOWER(name), created_at").Limit(q.Limit).Offset(q.Offset).Find(&apps).Error

	return apps, total, err
}

// Application returns one application by id.
func (s *Store) Application(ctx context.Context, id uuid.UUID) (*model.Application, error) {
	var app model.Application
	if err := s.db.WithContext(ctx).First(&app, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}

	return &app, nil
}

// ApplicationRoleCounts says how many roles each application defines, by id.
func (s *Store) ApplicationRoleCounts(ctx context.Context) (map[uuid.UUID]int64, error) {
	return s.countBy(ctx, "user_roles", "application_id", nil)
}

// CreateApplication writes a new application.
func (s *Store) CreateApplication(ctx context.Context, app *model.Application) error {
	return translate(s.db.WithContext(ctx).Create(app).Error)
}

// SaveApplication writes an application back.
func (s *Store) SaveApplication(ctx context.Context, app *model.Application) error {
	return translate(s.db.WithContext(ctx).Save(app).Error)
}

// DeleteApplication removes an application for good, with every role it
// defines, every user's hold on those roles and the API scopes they grant,
// every administrator's role scoped to it, and its access to APIs. The joins are removed by hand rather than left to the foreign
// keys, so this does not depend on how those constraints were created.
func (s *Store) DeleteApplication(ctx context.Context, app *model.Application) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		roles := "SELECT id FROM user_roles WHERE application_id = @id"

		statements := []string{
			"DELETE FROM user_role_members WHERE user_role_id IN (" + roles + ")",
			"DELETE FROM user_role_inherits WHERE role_id IN (" + roles + ") OR inherited_role_id IN (" + roles + ")",
			"DELETE FROM user_role_api_scopes WHERE user_role_id IN (" + roles + ")",
			"DELETE FROM user_roles WHERE application_id = @id",
			"DELETE FROM admin_role_assignments WHERE application_id = @id",
			"DELETE FROM application_api_scopes WHERE application_id = @id",
			"DELETE FROM application_apis WHERE application_id = @id",
		}

		for _, statement := range statements {
			if err := tx.Exec(statement, map[string]any{"id": app.ID}).Error; err != nil {
				return err
			}
		}

		return tx.Unscoped().Delete(app).Error
	})
}
