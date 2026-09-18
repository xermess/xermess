package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"xermess/internal/model"
)

func init() {
	goose.AddMigrationContext(upSchema, downSchema)
}

// upSchema builds the whole database in one step: every table the models
// describe, the indexes they cannot describe themselves, and the few rows an
// installation starts with.
//
// There is one migration on purpose. The schema is the models — `model.All()`
// is the list, and each struct says what its columns are — so a change is a
// change to a struct, and this file only has to keep saying "make what the
// models describe". A second migration would be worth writing when a database
// somewhere is already running and has to be brought forward without being
// rebuilt; until then, one file is the whole story and there is nowhere for
// the two to disagree.
//
// What it creates, in order, because a table has to exist before the tables
// that point at it: the organisation and the panel's own settings, the admin
// roles and the administrators holding them, the users and the fields their
// records are made of, the applications and the APIs they ask for tokens for,
// the provider's own tables — codes, tokens, sessions, keys — and last the
// providers users may sign in with and the identities they hold there.
func upSchema(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(model.All()...); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	if err := createIndexes(db); err != nil {
		return err
	}

	return seed(db)
}

// downSchema drops everything again, children and join tables before the
// tables they point at, so the foreign keys do not block it.
//
// It is the whole schema: there is nothing to step back to.
func downSchema(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}

	return db.Migrator().DropTable(
		// What signs users in from somewhere else.
		&model.SocialLogin{},
		&model.UserIdentity{},
		&model.SocialProvider{},

		// What the provider issued, and what it issued them for.
		&model.PasswordReset{},
		&model.UserSession{},
		&model.RefreshToken{},
		&model.AuthorizationCode{},
		&model.AuthorizationRequest{},
		&model.SigningKey{},

		// Who holds what.
		&model.AdminRoleAssignment{},
		"user_role_members",
		"user_role_inherits",
		"user_role_api_scopes",
		&model.ApplicationAPIScope{},
		&model.ApplicationAPI{},
		&model.APIScope{},
		&model.API{},

		// The accounts themselves, and everything about them.
		&model.MFA{},
		&model.AdminUserSession{},
		&model.AuditLog{},
		&model.AdminUser{},
		&model.Role{},
		&model.User{},
		&model.UserRole{},
		&model.Application{},
		&model.UserField{},

		// The settings.
		&model.Translation{},
		&model.Language{},
		&model.LoginFlow{},
		&model.AdminSecurity{},
		&model.Organization{},
	)
}

// createIndexes adds the indexes a struct tag cannot describe.
//
// A null application means the whole panel for an admin role, and a global
// role for a user role. Postgres treats every null as different, so a unique
// index over a nullable column needs a partial index for the null case.
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_global_name
			ON user_roles (name) WHERE application_id IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_role_assignments_global
			ON admin_role_assignments (admin_user_id, role_id) WHERE application_id IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_role_assignments_scoped
			ON admin_role_assignments (admin_user_id, role_id, application_id) WHERE application_id IS NOT NULL`,
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}

	return nil
}

// seed writes the rows an installation cannot start without: the admin roles
// there are to hold, the organisation the panel edits, the login flow every
// application falls back to, and the language every page falls back to.
//
// It seeds nothing else. There are no applications, no user fields and no
// user roles: what a record is made of and who may hold what is decided in
// the panel, so a new installation starts with none of it.
//
// There is no administrator either. The first one is made by whoever opens
// the panel, on /admin/new-super-admin, which is offered only while there is
// none — so a deployment has no password written down anywhere, and no
// default one to forget to change. How administrators are then made to sign
// in (admin_security) is written on the server's first start, from the
// configuration, for the same reason: it is a setting, not a schema.
//
// Everything here is safe to run twice.
func seed(db *gorm.DB) error {
	for _, starting := range startingRoles {
		role := starting

		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&role).Error; err != nil {
			return fmt.Errorf("create role %s: %w", starting.Name, err)
		}
	}

	var organizations int64
	if err := db.Model(&model.Organization{}).Count(&organizations).Error; err != nil {
		return fmt.Errorf("count organizations: %w", err)
	}

	if organizations == 0 {
		organization := model.DefaultOrganization()
		if err := db.Create(&organization).Error; err != nil {
			return fmt.Errorf("create the organization: %w", err)
		}
	}

	var flows int64
	if err := db.Model(&model.LoginFlow{}).Count(&flows).Error; err != nil {
		return fmt.Errorf("count login flows: %w", err)
	}

	if flows == 0 {
		flow := model.DefaultLoginFlow()
		if err := db.Create(&flow).Error; err != nil {
			return fmt.Errorf("create the default login flow: %w", err)
		}
	}

	// Only the base language's row. Its text, and the other languages the
	// server ships with, are imported on the first start by
	// store.EnsureLanguages rather than here: the files change with every
	// release, and a migration has to say the same thing forever.
	var languages int64
	if err := db.Model(&model.Language{}).Count(&languages).Error; err != nil {
		return fmt.Errorf("count languages: %w", err)
	}

	if languages == 0 {
		language := model.DefaultLanguage()
		if err := db.Create(&language).Error; err != nil {
			return fmt.Errorf("create the default language: %w", err)
		}
	}

	return nil
}

// startingRoles are the admin roles a panel is given, and what each grants.
//
// super_admin is built in: it grants every permission by name, so it lists
// none, and it is the only role that can manage administrators. A super admin
// can change or remove any of the others. app_manager is meant to be held for
// one application: every one of its permissions can be scoped.
var startingRoles = []model.Role{
	{
		Name:        model.RoleSuperAdmin,
		Description: "Everything, including managing administrators",
	},
	{
		Name:        "admin",
		Description: "Everything but managing administrators",
		Permissions: model.AdminPermissionNames(),
	},
	{
		Name:        "moderator",
		Description: "Manage users, application roles and who holds them, and read the activity log",
		Permissions: []string{
			model.PermActivityRead, model.PermUsersRead, model.PermUsersWrite, model.PermAPIsRead,
			model.PermApplicationsRead, model.PermUserRolesWrite, model.PermRoleAssignmentsWrite,
		},
	},
	{
		Name:        "app_manager",
		Description: "Look after an application: its settings, its roles, and who holds them",
		Permissions: []string{
			model.PermApplicationsRead, model.PermApplicationsWrite,
			model.PermUserRolesWrite, model.PermRoleAssignmentsWrite,
		},
	},
	{
		Name:        "support",
		Description: "Look users up, fix their records and give them application roles",
		Permissions: []string{
			model.PermUsersRead, model.PermUsersWrite, model.PermApplicationsRead, model.PermRoleAssignmentsWrite,
		},
	},
	{
		Name:        "auditor",
		Description: "Read only, including the activity log",
		Permissions: []string{
			model.PermActivityRead, model.PermUsersRead, model.PermAPIsRead, model.PermApplicationsRead,
		},
	},
}
