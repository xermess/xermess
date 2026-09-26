package store

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"loginer/internal/model"
)

// ---- Second factors -------------------------------------------------------

// MFAFactors returns an administrator's factors of one method, confirmed and
// not, newest first.
func (s *Store) MFAFactors(ctx context.Context, admin uuid.UUID, method model.MFAMethod) ([]model.MFA, error) {
	var factors []model.MFA
	err := s.db.WithContext(ctx).
		Where("admin_user_id = ? AND method = ?", admin, method).
		Order("created_at DESC").
		Find(&factors).Error

	return factors, err
}

// StartMFA stores a new, unconfirmed factor, removing any other unconfirmed one
// of the same method: only the enrolment under way can be confirmed.
func (s *Store) StartMFA(ctx context.Context, factor *model.MFA) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().
			Where("admin_user_id = ? AND method = ? AND confirmed_at IS NULL", factor.AdminUserID, factor.Method).
			Delete(&model.MFA{}).Error
		if err != nil {
			return err
		}

		return translate(tx.Omit(clause.Associations).Create(factor).Error)
	})
}

// ConfirmMFA marks a factor confirmed with its recovery codes, and removes every
// other factor of the same method: confirming a new authenticator replaces the
// old one.
func (s *Store) ConfirmMFA(ctx context.Context, factor *model.MFA) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().
			Where("admin_user_id = ? AND method = ? AND id <> ?", factor.AdminUserID, factor.Method, factor.ID).
			Delete(&model.MFA{}).Error
		if err != nil {
			return err
		}

		return tx.Omit(clause.Associations).Save(factor).Error
	})
}

// ClaimMFAStep records that a factor accepted a code for the step starting at
// `at`. Only the first of two requests presenting codes for the same step
// succeeds, so a code cannot be used twice even at the same moment.
func (s *Store) ClaimMFAStep(ctx context.Context, factor uuid.UUID, at time.Time) (bool, error) {
	result := s.db.WithContext(ctx).Model(&model.MFA{}).
		Where("id = ? AND (last_used_at IS NULL OR last_used_at < ?)", factor, at).
		Update("last_used_at", at)

	return result.RowsAffected == 1, result.Error
}

// ConsumeRecoveryCode removes one recovery code from a factor, if it holds it,
// and says whether it did. It locks the row, so a code works once even when
// presented twice at once.
//
// It leaves LastUsedAt alone. That column is the start of the last TOTP step
// accepted (model.MFA.UsedStep), not a note of when the factor was last used
// at all: writing the wall clock into it would make the step just gone look
// spent, and the code on the administrator's phone would be refused until the
// next one appeared. What was used, and when, is in the activity log.
func (s *Store) ConsumeRecoveryCode(ctx context.Context, factor uuid.UUID, hash string) (bool, error) {
	consumed := false

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.MFA
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", factor).Error; err != nil {
			return translate(err)
		}

		index := slices.Index(row.RecoveryCodes, hash)
		if index < 0 {
			return nil
		}

		consumed = true
		// Through the struct, with the column named, so the codes go through
		// their JSON serializer; a plain column update would not.
		return tx.Model(&row).Select("RecoveryCodes").Updates(&model.MFA{
			RecoveryCodes: slices.Delete(slices.Clone(row.RecoveryCodes), index, index+1),
		}).Error
	})

	return consumed, err
}

// SetRecoveryCodes replaces a factor's recovery codes.
func (s *Store) SetRecoveryCodes(ctx context.Context, factor *model.MFA) error {
	return s.db.WithContext(ctx).Model(factor).Select("RecoveryCodes").Updates(&model.MFA{RecoveryCodes: factor.RecoveryCodes}).Error
}

// RemoveMFA deletes every factor an administrator has and ends their sessions:
// what turning two-factor sign-in off, or a super admin resetting it, means.
// With `keep` set, that one session stays signed in.
func (s *Store) RemoveMFA(ctx context.Context, admin uuid.UUID, keep *uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("admin_user_id = ?", admin).Delete(&model.MFA{}).Error; err != nil {
			return err
		}

		query := tx.Model(&model.AdminUserSession{}).Where("admin_user_id = ? AND revoked_at IS NULL", admin)
		if keep != nil {
			query = query.Where("id <> ?", *keep)
		}

		return query.Update("revoked_at", at).Error
	})
}

// PassSessionMFA marks a session's second factor cleared, and gives it the
// lifetime of a full session from now.
func (s *Store) PassSessionMFA(ctx context.Context, session uuid.UUID, expires time.Time) error {
	return s.db.WithContext(ctx).Model(&model.AdminUserSession{}).
		Where("id = ? AND revoked_at IS NULL", session).
		Updates(map[string]any{"mfa_passed": true, "expires_at": expires}).Error
}
