package store

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"loginer/internal/model"
)

// scopesByName orders an API's scopes, so they always come out the same way.
func scopesByName(db *gorm.DB) *gorm.DB {
	return db.Order("name")
}

// APIs returns every API matching a search, sorted by name, with its scopes.
// There are few enough of them that the panel shows all at once.
func (s *Store) APIs(ctx context.Context, search string) ([]model.API, error) {
	query := s.db.WithContext(ctx).Model(&model.API{}).Preload("Scopes", scopesByName)

	if search = strings.TrimSpace(search); search != "" {
		like := contains(search)
		query = query.Where("LOWER(name) LIKE ? OR LOWER(identifier) LIKE ? OR LOWER(description) LIKE ?", like, like, like)
	}

	var apis []model.API
	err := query.Order("LOWER(name)").Find(&apis).Error

	return apis, err
}

// API returns one API by id, with its scopes.
func (s *Store) API(ctx context.Context, id uuid.UUID) (*model.API, error) {
	var api model.API
	if err := s.db.WithContext(ctx).Preload("Scopes", scopesByName).First(&api, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}

	return &api, nil
}

// APIByIdentifier returns the API a token audience names, with its scopes.
func (s *Store) APIByIdentifier(ctx context.Context, identifier string) (*model.API, error) {
	var api model.API
	err := s.db.WithContext(ctx).Preload("Scopes", scopesByName).First(&api, "identifier = ?", identifier).Error
	if err != nil {
		return nil, translate(err)
	}

	return &api, nil
}

// APIApplicationCounts says how many applications may use each API, by id.
func (s *Store) APIApplicationCounts(ctx context.Context) (map[uuid.UUID]int64, error) {
	return s.countBy(ctx, "application_apis", "api_id", nil)
}

// APIRoleCounts says how many roles grant at least one scope of each API, by
// id.
func (s *Store) APIRoleCounts(ctx context.Context) (map[uuid.UUID]int64, error) {
	var rows []struct {
		ID    uuid.UUID
		Count int64
	}

	err := s.db.WithContext(ctx).
		Table("user_role_api_scopes").
		Joins("JOIN api_scopes ON api_scopes.id = user_role_api_scopes.api_scope_id").
		Select("api_scopes.api_id AS id, COUNT(DISTINCT user_role_api_scopes.user_role_id) AS count").
		Group("api_scopes.api_id").
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

// APIApplication is what one application may do with an API.
type APIApplication struct {
	Application model.Application
	Authorized  bool
	// Allowed are the ids of the API's scopes the application may ask for.
	Allowed []uuid.UUID
}

// APIApplications lists applications with what each may do with an API,
// sorted by name. `only` narrows them to the applications an administrator
// can see; nil is every one.
func (s *Store) APIApplications(ctx context.Context, apiID uuid.UUID, only []uuid.UUID) ([]APIApplication, error) {
	apps, _, err := s.Applications(ctx, ApplicationQuery{Only: only, Limit: 1000})
	if err != nil {
		return nil, err
	}

	var authorized []uuid.UUID
	if err := s.db.WithContext(ctx).Model(&model.ApplicationAPI{}).
		Where("api_id = ?", apiID).Pluck("application_id", &authorized).Error; err != nil {
		return nil, err
	}

	var rows []model.ApplicationAPIScope
	err = s.db.WithContext(ctx).
		Where("api_scope_id IN (SELECT id FROM api_scopes WHERE api_id = ?)", apiID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]APIApplication, 0, len(apps))
	for _, app := range apps {
		entry := APIApplication{Application: app, Authorized: slices.Contains(authorized, app.ID), Allowed: []uuid.UUID{}}
		for _, row := range rows {
			if row.ApplicationID == app.ID {
				entry.Allowed = append(entry.Allowed, row.APIScopeID)
			}
		}
		out = append(out, entry)
	}

	return out, nil
}

// APIAuditLog returns the newest entries involving an API: changes to the API
// itself, and applications being authorised for it or losing access.
func (s *Store) APIAuditLog(ctx context.Context, apiID uuid.UUID, limit int) ([]model.AuditLog, error) {
	var events []model.AuditLog
	err := s.db.WithContext(ctx).
		Where("(target_type = ? AND target_id = ?) OR (metadata IS NOT NULL AND metadata::jsonb ->> 'api_id' = ?)",
			"api", apiID.String(), apiID.String()).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error

	return events, err
}

// APIScopesByID returns the API scopes with these ids. An id named twice is
// one scope, and an id with no scope behind it is ErrNotFound.
func (s *Store) APIScopesByID(ctx context.Context, ids []uuid.UUID) ([]model.APIScope, error) {
	scopes := []model.APIScope{}
	if err := s.byID(ctx, ids, &scopes, func() int { return len(scopes) }); err != nil {
		return nil, err
	}

	return scopes, nil
}

// CreateAPI writes a new API with its scopes, in one transaction.
func (s *Store) CreateAPI(ctx context.Context, api *model.API) error {
	return translate(s.db.WithContext(ctx).Create(api).Error)
}

// SaveAPI writes an API back and makes its scopes the ones it carries. A scope
// that keeps its id keeps every application's allowance and role's grant of
// it, whatever it is renamed to; a scope left out is removed with them.
func (s *Store) SaveAPI(ctx context.Context, api *model.API) error {
	return translate(s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Save(api).Error; err != nil {
			return err
		}

		kept := []uuid.UUID{}
		for _, scope := range api.Scopes {
			if scope.ID != uuid.Nil {
				kept = append(kept, scope.ID)
			}
		}

		gone := tx.Model(&model.APIScope{}).Select("id").Where("api_id = ?", api.ID)
		if len(kept) > 0 {
			gone = gone.Where("id NOT IN ?", kept)
		}

		var goneIDs []uuid.UUID
		if err := gone.Pluck("id", &goneIDs).Error; err != nil {
			return err
		}

		if err := removeScopes(tx, goneIDs); err != nil {
			return err
		}

		// The unique index on a scope's name is checked row by row, so two
		// scopes trading names would clash halfway through. The kept scopes
		// are moved out of the way first, onto their own ids, which no scope
		// can be called.
		if len(kept) > 0 {
			err := tx.Exec("UPDATE api_scopes SET name = id::text WHERE api_id = ? AND id IN ?", api.ID, kept).Error
			if err != nil {
				return err
			}
		}

		for i := range api.Scopes {
			scope := &api.Scopes[i]
			scope.APIID = api.ID

			if scope.ID == uuid.Nil {
				if err := tx.Create(scope).Error; err != nil {
					return err
				}
				continue
			}

			err := tx.Model(&model.APIScope{}).
				Where("id = ? AND api_id = ?", scope.ID, api.ID).
				Updates(map[string]any{
					"name":        scope.Name,
					"description": scope.Description,
					"default":     scope.Default,
				}).Error
			if err != nil {
				return err
			}
		}

		return nil
	}))
}

// DeleteAPI removes an API for good, with its scopes, every application's
// authorisation for it, and every role's grant of its scopes.
func (s *Store) DeleteAPI(ctx context.Context, api *model.API) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var scopeIDs []uuid.UUID
		if err := tx.Model(&model.APIScope{}).Where("api_id = ?", api.ID).Pluck("id", &scopeIDs).Error; err != nil {
			return err
		}

		if err := removeScopes(tx, scopeIDs); err != nil {
			return err
		}

		if err := tx.Exec("DELETE FROM application_apis WHERE api_id = ?", api.ID).Error; err != nil {
			return err
		}

		return tx.Unscoped().Delete(api).Error
	})
}

// removeScopes deletes API scopes and every allowance and grant of them.
func removeScopes(tx *gorm.DB, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	statements := []string{
		"DELETE FROM application_api_scopes WHERE api_scope_id IN ?",
		"DELETE FROM user_role_api_scopes WHERE api_scope_id IN ?",
		"DELETE FROM api_scopes WHERE id IN ?",
	}

	for _, statement := range statements {
		if err := tx.Exec(statement, ids).Error; err != nil {
			return err
		}
	}

	return nil
}

// APIAccess is what one application may do with one API: whether it may ask
// for tokens for it at all, and which of its scopes.
type APIAccess struct {
	API        model.API
	Authorized bool
	// Allowed are the ids of the API's scopes the application may ask for.
	Allowed []uuid.UUID
}

// ApplicationAPIAccess lists every API with what the application may do with
// it, sorted by API name.
func (s *Store) ApplicationAPIAccess(ctx context.Context, applicationID uuid.UUID) ([]APIAccess, error) {
	apis, err := s.APIs(ctx, "")
	if err != nil {
		return nil, err
	}

	var authorized []uuid.UUID
	err = s.db.WithContext(ctx).Model(&model.ApplicationAPI{}).
		Where("application_id = ?", applicationID).Pluck("api_id", &authorized).Error
	if err != nil {
		return nil, err
	}

	var allowed []uuid.UUID
	err = s.db.WithContext(ctx).Model(&model.ApplicationAPIScope{}).
		Where("application_id = ?", applicationID).Pluck("api_scope_id", &allowed).Error
	if err != nil {
		return nil, err
	}

	out := make([]APIAccess, 0, len(apis))
	for _, api := range apis {
		access := APIAccess{API: api, Allowed: []uuid.UUID{}}

		for _, id := range authorized {
			if id == api.ID {
				access.Authorized = true
			}
		}

		for _, scope := range api.Scopes {
			for _, id := range allowed {
				if id == scope.ID {
					access.Allowed = append(access.Allowed, scope.ID)
				}
			}
		}

		out = append(out, access)
	}

	return out, nil
}

// AuthorizeApplicationAPI lets an application ask for tokens for an API, and
// makes `scopes` — which must be the API's — the scopes it may ask for.
func (s *Store) AuthorizeApplicationAPI(ctx context.Context, applicationID, apiID uuid.UUID, scopes []uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		grant := model.ApplicationAPI{ApplicationID: applicationID, APIID: apiID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&grant).Error; err != nil {
			return err
		}

		err := tx.Exec(
			"DELETE FROM application_api_scopes WHERE application_id = ? AND api_scope_id IN (SELECT id FROM api_scopes WHERE api_id = ?)",
			applicationID, apiID,
		).Error
		if err != nil {
			return err
		}

		for _, scope := range scopes {
			row := model.ApplicationAPIScope{ApplicationID: applicationID, APIScopeID: scope}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// RevokeApplicationAPI stops an application asking for tokens for an API.
func (s *Store) RevokeApplicationAPI(ctx context.Context, applicationID, apiID uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Exec(
			"DELETE FROM application_api_scopes WHERE application_id = ? AND api_scope_id IN (SELECT id FROM api_scopes WHERE api_id = ?)",
			applicationID, apiID,
		).Error
		if err != nil {
			return err
		}

		return tx.Exec("DELETE FROM application_apis WHERE application_id = ? AND api_id = ?", applicationID, apiID).Error
	})
}
