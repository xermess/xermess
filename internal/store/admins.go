package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"loginer/internal/brand"
	"loginer/internal/cache"
	"loginer/internal/model"
)

// ErrAdminExists is returned when the first administrator is asked for and
// there already is one. It is what keeps the endpoint that creates it from
// being a way in once the panel is set up.
var ErrAdminExists = errors.New("an administrator already exists")

// firstAdminLock names the advisory lock CreateFirstAdmin holds. Any number
// will do, as long as nothing else locks the same one.
const firstAdminLock = brand.AdminLock

// AdminsExist reports whether anyone can sign in to the panel yet.
func (s *Store) AdminsExist(ctx context.Context) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Admin{}).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateFirstAdmin writes the administrator a new installation is set up
// with, and refuses if there is one already.
//
// The check and the write are one transaction under a lock, so two people
// submitting the setup form at the same moment cannot both get an account:
// the second waits for the first, finds it, and is turned away.
func (s *Store) CreateFirstAdmin(ctx context.Context, admin *model.Admin) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Counting alone does not stop two transactions that start together:
		// each would count none and write its own. A lock held until the
		// transaction ends makes the second one wait, then count the first.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", firstAdminLock).Error; err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&model.Admin{}).Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return ErrAdminExists
		}

		var role model.AdminRole
		if err := tx.Where("name = ?", model.RoleSuperAdmin).First(&role).Error; err != nil {
			return err
		}

		admin.Assignments = nil
		if err := translate(tx.Omit(clause.Associations).Create(admin).Error); err != nil {
			return err
		}

		assignment := model.AdminRoleAssignment{AdminID: admin.ID, RoleID: role.ID, Role: role}
		if err := tx.Omit(clause.Associations).Create(&assignment).Error; err != nil {
			return err
		}

		admin.Assignments = []model.AdminRoleAssignment{assignment}

		return nil
	})
}

// withAssignments loads the roles an administrator holds, each with its role
// and the application it is scoped to, whole-panel ones first — and their
// confirmed second factors, so whether they have two-factor sign-in is known
// without another query.
func withAssignments(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Assignments", func(db *gorm.DB) *gorm.DB {
			return db.Order("application_id NULLS FIRST, created_at")
		}).
		Preload("Assignments.Role").
		Preload("Assignments.Application").
		Preload("MFA", "confirmed_at IS NOT NULL")
}

// AdminByUsername loads an administrator and the roles they hold. It is what
// signing in starts with.
func (s *Store) AdminByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	err := withAssignments(s.db.WithContext(ctx)).
		Where("username = ?", username).
		First(&admin).Error
	if err != nil {
		return nil, translate(err)
	}

	return &admin, nil
}

// AdminByID loads an administrator with the roles they hold and where they
// hold them, which is what the administrator is allowed to do: every request
// is checked against this.
func (s *Store) AdminByID(ctx context.Context, id uuid.UUID) (*model.Admin, error) {
	var admin model.Admin
	err := withAssignments(s.db.WithContext(ctx)).
		First(&admin, "id = ?", id).Error
	if err != nil {
		return nil, translate(err)
	}

	return &admin, nil
}

// principal is how an administrator is kept in the session database: the
// account, the roles they hold and where, and their confirmed second factors
// — spelled out, because the model leaves the roles and the factors out of
// its JSON. The password hash and the factors' secrets are not kept.
type principal struct {
	Admin       model.Admin                 `json:"admin"`
	Assignments []model.AdminRoleAssignment `json:"assignments"`
	Factors     []model.MFA                 `json:"factors"`
}

// AdminPrincipal is AdminByID for deciding what a signed-in request may do,
// read through the session database: every request the panel makes asks.
//
// What it answers has no password hash, so it is never written back and never
// checked a password against — load the administrator with AdminByID for
// that.
func (s *Store) AdminPrincipal(ctx context.Context, id uuid.UUID) (*model.Admin, error) {
	kept, err := cached(ctx, s, cache.Admins, "id:"+id.String(), func() (principal, error) {
		admin, err := s.AdminByID(ctx, id)
		if err != nil {
			return principal{}, err
		}

		return principal{Admin: *admin, Assignments: admin.Assignments, Factors: admin.MFA}, nil
	})
	if err != nil {
		return nil, err
	}

	admin := kept.Admin
	admin.Assignments = kept.Assignments
	admin.MFA = kept.Factors

	return &admin, nil
}

// MarkAdminSignedIn records when and from where an administrator last signed
// in, and forgets the wrong passwords that came before.
func (s *Store) MarkAdminSignedIn(ctx context.Context, admin *model.Admin, at time.Time, ip string) error {
	err := s.db.WithContext(ctx).Model(admin).Updates(map[string]any{
		"last_login_at":      at,
		"last_login_ip":      ip,
		"failed_login_count": 0,
		"locked_until":       nil,
	}).Error
	if err == nil {
		s.forget(ctx, cache.Admins)
	}

	return err
}

// RecordFailedLogin counts a wrong password against an administrator and, at
// the `max`th in a row, locks the account until `lockFor` from now and starts
// counting again. It reports whether this attempt locked the account.
//
// The count is added to in the database rather than read and written back,
// so attempts arriving at once are all counted.
func (s *Store) RecordFailedLogin(ctx context.Context, admin *model.Admin, at time.Time, max int, lockFor time.Duration) (bool, error) {
	var row struct {
		LockedUntil *time.Time
	}

	err := s.db.WithContext(ctx).Raw(`
		UPDATE admins SET
			locked_until = CASE WHEN failed_login_count + 1 >= @max THEN @until ELSE locked_until END,
			failed_login_count = CASE WHEN failed_login_count + 1 >= @max THEN 0 ELSE failed_login_count + 1 END
		WHERE id = @id
		RETURNING locked_until`,
		map[string]any{"max": max, "until": at.Add(lockFor), "id": admin.ID},
	).Scan(&row).Error
	if err != nil {
		return false, err
	}

	// A lock has to reach the session database, or a request already signed
	// in would go on as if the account were open.
	s.forget(ctx, cache.Admins)

	return row.LockedUntil != nil && row.LockedUntil.After(at), nil
}

// AdminQuery is what a listing of administrators asks for: a search box, a
// filter, and a page.
type AdminQuery struct {
	// Search matches the username, the email or the name.
	Search string
	// Status filters on the account's status. Empty means every status.
	Status model.Status
	// Role keeps the administrators who hold this role, anywhere. Nil means
	// everyone.
	Role   *uuid.UUID
	Limit  int
	Offset int
}

// Admins returns a page of administrators, newest first, with their roles,
// along with how many match the query in total.
func (s *Store) Admins(ctx context.Context, q AdminQuery) ([]model.Admin, int64, error) {
	query := s.db.WithContext(ctx).Model(&model.Admin{})

	if search := strings.TrimSpace(q.Search); search != "" {
		like := contains(search)
		query = query.Where(
			`LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?`,
			like, like, like, like,
		)
	}

	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
	}

	if q.Role != nil {
		query = query.Where("id IN (SELECT admin_id FROM admin_role_assignments WHERE role_id = ?)", *q.Role)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var admins []model.Admin
	err := withAssignments(query).
		Order("created_at DESC").
		Limit(q.Limit).
		Offset(q.Offset).
		Find(&admins).Error

	return admins, total, err
}

// CreateAdmin writes a new administrator with the roles they carry, in one
// transaction.
func (s *Store) CreateAdmin(ctx context.Context, admin *model.Admin) error {
	return translate(s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(admin).Error; err != nil {
			return err
		}

		return writeAssignments(tx, admin)
	}))
}

// SaveAdmin writes an administrator back and replaces the roles they hold with
// the ones they carry, in one transaction.
func (s *Store) SaveAdmin(ctx context.Context, admin *model.Admin) error {
	return s.forgetting(ctx, translate(s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Save(admin).Error; err != nil {
			return err
		}

		if err := tx.Where("admin_id = ?", admin.ID).Delete(&model.AdminRoleAssignment{}).Error; err != nil {
			return err
		}

		return writeAssignments(tx, admin)
	})), cache.Admins)
}

// SaveOwnAccount writes the columns an administrator may change about
// themselves — their name, their address, their password — and nothing
// else, so a change made from their own profile can never reach the roles
// they hold. A taken address is ErrDuplicate.
func (s *Store) SaveOwnAccount(ctx context.Context, admin *model.Admin) error {
	return s.forgetting(ctx, translate(s.db.WithContext(ctx).
		Model(admin).
		Select("first_name", "last_name", "email", "username", "avatar_url", "password_hash").
		Updates(admin).Error), cache.Admins)
}

// writeAssignments inserts the administrator's assignments. The roles and
// applications they point at already exist, so only the assignments are
// written.
func writeAssignments(tx *gorm.DB, admin *model.Admin) error {
	for i := range admin.Assignments {
		admin.Assignments[i].ID = uuid.Nil
		admin.Assignments[i].AdminID = admin.ID

		if err := tx.Omit(clause.Associations).Create(&admin.Assignments[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

// DeleteAdmin removes an administrator for good, with their roles, sessions
// and second factors. The activity log is left as it is: its rows name the
// actor by username as well as by id, so what they did stays readable.
func (s *Store) DeleteAdmin(ctx context.Context, admin *model.Admin) error {
	return s.forgetting(ctx, s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		statements := []string{
			"DELETE FROM admin_role_assignments WHERE admin_id = @id",
			"DELETE FROM admin_sessions WHERE admin_id = @id",
			"DELETE FROM mfa_factors WHERE admin_id = @id",
		}

		for _, statement := range statements {
			if err := tx.Exec(statement, map[string]any{"id": admin.ID}).Error; err != nil {
				return err
			}
		}

		return tx.Unscoped().Delete(admin).Error
	}), cache.Admins)
}
