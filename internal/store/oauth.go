package store

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"xermess/internal/model"
)

// ErrAlreadyUsed is returned when a single-use secret — an authorization code,
// a refresh token, a reset link — is presented after it was used. It is told
// apart from ErrNotFound because the provider answers a replay by revoking
// what the first use issued.
var ErrAlreadyUsed = errors.New("already used")

// ---- Signing keys ---------------------------------------------------------

// SigningKeys returns every stored signing key, newest first.
func (s *Store) SigningKeys(ctx context.Context) ([]model.SigningKey, error) {
	var keys []model.SigningKey
	err := s.db.WithContext(ctx).Order("created_at DESC").Find(&keys).Error

	return keys, err
}

// CreateSigningKey stores a new signing key.
func (s *Store) CreateSigningKey(ctx context.Context, key *model.SigningKey) error {
	return translate(s.db.WithContext(ctx).Create(key).Error)
}

// ---- Applications and audiences -------------------------------------------

// ApplicationByClientID returns the application a client id belongs to.
func (s *Store) ApplicationByClientID(ctx context.Context, clientID string) (*model.Application, error) {
	var app model.Application
	if err := s.db.WithContext(ctx).First(&app, "client_id = ?", clientID).Error; err != nil {
		return nil, translate(err)
	}

	return &app, nil
}

// Audience is what a token request naming an API needs to know about it: the
// API, whether the application may ask for tokens for it, and the names of the
// scopes it may ask for.
type Audience struct {
	API        *model.API
	Authorized bool
	Allowed    []string
}

// AudienceFor loads the API with this identifier and what the application may
// do with it. An unknown identifier is ErrNotFound.
func (s *Store) AudienceFor(ctx context.Context, applicationID uuid.UUID, identifier string) (Audience, error) {
	api, err := s.APIByIdentifier(ctx, identifier)
	if err != nil {
		return Audience{}, err
	}

	access, err := s.ApplicationAPIAccess(ctx, applicationID)
	if err != nil {
		return Audience{}, err
	}

	out := Audience{API: api, Allowed: []string{}}
	for _, it := range access {
		if it.API.ID != api.ID {
			continue
		}

		out.Authorized = it.Authorized
		for _, scope := range api.Scopes {
			if slices.Contains(it.Allowed, scope.ID) {
				out.Allowed = append(out.Allowed, scope.Name)
			}
		}
	}

	return out, nil
}

// EffectiveRoles returns every role the user holds, inheritance followed, each
// with the API scopes it grants: what model.EvaluateToken is given.
func (s *Store) EffectiveRoles(ctx context.Context, user *model.User) ([]model.UserRole, error) {
	graph, err := s.RoleGraph(ctx)
	if err != nil {
		return nil, err
	}

	held := make([]uuid.UUID, 0, len(user.Roles))
	for _, role := range user.Roles {
		held = append(held, role.ID)
	}

	return graph.Effective(held), nil
}

// ---- Users signing in -----------------------------------------------------

// UserByEmail returns the user with this address, compared without case.
// Addresses are stored normalized (model.NormalizeEmail), so this is one
// lookup on the unique index rather than a scan of every user.
func (s *Store) UserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := s.db.WithContext(ctx).Preload("Roles", byName).First(&user, "email = ?", model.NormalizeEmail(email)).Error
	if err != nil {
		return nil, translate(err)
	}

	return &user, nil
}

// MarkUserSignedIn records when a user signed in, and forgets the wrong
// passwords that came before.
//
// It writes these columns and nothing else: GORM would otherwise write back
// the roles the user was loaded with, putting back any that were taken away
// during the sign-in — which is what single sign-on's role sync does.
func (s *Store) MarkUserSignedIn(ctx context.Context, user *model.User, at time.Time) error {
	return s.db.WithContext(ctx).Model(user).Omit(clause.Associations).Updates(map[string]any{
		"last_login_at":      at,
		"failed_login_count": 0,
		"locked_until":       nil,
	}).Error
}

// RecordUserFailedLogin counts a wrong password against a user and, at the
// `max`th in a row, locks the account until `lockFor` from now. It reports
// whether this attempt locked it. It is RecordFailedLogin for users.
func (s *Store) RecordUserFailedLogin(ctx context.Context, user *model.User, at time.Time, max int, lockFor time.Duration) (bool, error) {
	var row struct {
		LockedUntil *time.Time
	}

	err := s.db.WithContext(ctx).Raw(`
		UPDATE users SET
			locked_until = CASE WHEN failed_login_count + 1 >= @max THEN @until ELSE locked_until END,
			failed_login_count = CASE WHEN failed_login_count + 1 >= @max THEN 0 ELSE failed_login_count + 1 END
		WHERE id = @id
		RETURNING locked_until`,
		map[string]any{"max": max, "until": at.Add(lockFor), "id": user.ID},
	).Scan(&row).Error
	if err != nil {
		return false, err
	}

	return row.LockedUntil != nil && row.LockedUntil.After(at), nil
}

// ---- Authorization requests and codes -------------------------------------

// CreateAuthorizationRequest stores a sign-in under way.
func (s *Store) CreateAuthorizationRequest(ctx context.Context, req *model.AuthorizationRequest) error {
	return translate(s.db.WithContext(ctx).Omit(clause.Associations).Create(req).Error)
}

// AuthorizationRequestByHandle returns a sign-in under way, with its
// application.
func (s *Store) AuthorizationRequestByHandle(ctx context.Context, hash string) (*model.AuthorizationRequest, error) {
	var req model.AuthorizationRequest
	err := s.db.WithContext(ctx).Preload("Application").First(&req, "handle_hash = ?", hash).Error
	if err != nil {
		return nil, translate(err)
	}

	return &req, nil
}

// CompleteAuthorizationRequest marks a sign-in finished. Only the first call
// succeeds, so one handle can never be turned into two codes; the others get
// ErrAlreadyUsed.
func (s *Store) CompleteAuthorizationRequest(ctx context.Context, req *model.AuthorizationRequest, at time.Time) error {
	result := s.db.WithContext(ctx).Model(&model.AuthorizationRequest{}).
		Where("id = ? AND completed_at IS NULL", req.ID).
		Update("completed_at", at)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAlreadyUsed
	}

	req.CompletedAt = &at

	return nil
}

// CreateAuthorizationCode stores a code.
func (s *Store) CreateAuthorizationCode(ctx context.Context, code *model.AuthorizationCode) error {
	return translate(s.db.WithContext(ctx).Omit(clause.Associations).Create(code).Error)
}

// ClaimAuthorizationCode marks an application's own code used and returns it.
// Of two requests presenting the same code at once, one claims it and the
// other gets ErrAlreadyUsed — with the code too, so its tokens can be revoked.
//
// A code belonging to another application is never written to, and comes back
// with ErrAlreadyUsed as well; the caller tells the two apart by the code's
// ApplicationID. Claiming on the hash alone would let any client that had seen
// a code spend it out of the owner's hands by presenting it once.
func (s *Store) ClaimAuthorizationCode(ctx context.Context, hash string, application uuid.UUID, at time.Time) (*model.AuthorizationCode, error) {
	var codes []model.AuthorizationCode
	err := s.db.WithContext(ctx).Model(&codes).
		Clauses(clause.Returning{}).
		Where("code_hash = ? AND application_id = ? AND used_at IS NULL", hash, application).
		Update("used_at", at).Error
	if err != nil {
		return nil, err
	}
	if len(codes) == 1 {
		return &codes[0], nil
	}

	var used model.AuthorizationCode
	if err := s.db.WithContext(ctx).First(&used, "code_hash = ?", hash).Error; err != nil {
		return nil, translate(err)
	}

	return &used, ErrAlreadyUsed
}

// ---- Refresh tokens -------------------------------------------------------

// CreateRefreshToken stores a refresh token.
func (s *Store) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	return translate(s.db.WithContext(ctx).Omit(clause.Associations).Create(token).Error)
}

// RefreshTokenByHash returns a refresh token.
func (s *Store) RefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	if err := s.db.WithContext(ctx).First(&token, "token_hash = ?", hash).Error; err != nil {
		return nil, translate(err)
	}

	return &token, nil
}

// RotateRefreshToken revokes `old` and stores `next` in its place, in one
// transaction. If `old` was already revoked — by a rotation that got there
// first, or because it was stolen and used — nothing is stored and the answer
// is ErrAlreadyUsed.
func (s *Store) RotateRefreshToken(ctx context.Context, old, next *model.RefreshToken, at time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.RefreshToken{}).
			Where("id = ? AND revoked_at IS NULL", old.ID).
			Update("revoked_at", at)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrAlreadyUsed
		}

		return translate(tx.Omit(clause.Associations).Create(next).Error)
	})
}

// RevokeRefreshFamily revokes every token descended from the same grant.
func (s *Store) RevokeRefreshFamily(ctx context.Context, family uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("family_id = ? AND revoked_at IS NULL", family).
		Update("revoked_at", at).Error
}

// RevokeRefreshTokensForCode revokes what an authorization code was exchanged
// for, when the code is presented a second time.
func (s *Store) RevokeRefreshTokensForCode(ctx context.Context, code uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("code_id = ? AND revoked_at IS NULL", code).
		Update("revoked_at", at).Error
}

// RevokeRefreshTokensForUser revokes every refresh token a user has, in every
// application: what a password reset does.
func (s *Store) RevokeRefreshTokensForUser(ctx context.Context, user uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", user).
		Update("revoked_at", at).Error
}

// ---- User sessions --------------------------------------------------------

// CreateUserSession starts a user session.
func (s *Store) CreateUserSession(ctx context.Context, session *model.UserSession) error {
	return translate(s.db.WithContext(ctx).Omit(clause.Associations).Create(session).Error)
}

// UserSessionByHash returns the session a cookie belongs to.
func (s *Store) UserSessionByHash(ctx context.Context, hash string) (*model.UserSession, error) {
	var session model.UserSession
	if err := s.db.WithContext(ctx).First(&session, "token_hash = ?", hash).Error; err != nil {
		return nil, translate(err)
	}

	return &session, nil
}

// RevokeUserSession ends a session.
func (s *Store) RevokeUserSession(ctx context.Context, id uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).Model(&model.UserSession{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", at).Error
}

// RevokeUserSessionsFor ends every session a user has, in every browser.
func (s *Store) RevokeUserSessionsFor(ctx context.Context, user uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).Model(&model.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", user).
		Update("revoked_at", at).Error
}

// ---- Password resets ------------------------------------------------------

// CreatePasswordReset stores a reset link that is about to be sent.
func (s *Store) CreatePasswordReset(ctx context.Context, reset *model.PasswordReset) error {
	return translate(s.db.WithContext(ctx).Omit(clause.Associations).Create(reset).Error)
}

// PasswordResetByHash returns a reset link.
func (s *Store) PasswordResetByHash(ctx context.Context, hash string) (*model.PasswordReset, error) {
	var reset model.PasswordReset
	if err := s.db.WithContext(ctx).First(&reset, "token_hash = ?", hash).Error; err != nil {
		return nil, translate(err)
	}

	return &reset, nil
}

// ResetPassword uses a reset link: it marks the link used, saves the user's
// new password, and ends every session and refresh token the user has, in one
// transaction. A link used by someone else first is ErrAlreadyUsed.
func (s *Store) ResetPassword(ctx context.Context, reset *model.PasswordReset, user *model.User, at time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.PasswordReset{}).
			Where("id = ? AND used_at IS NULL", reset.ID).
			Update("used_at", at)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrAlreadyUsed
		}

		// The link was opened from the address's own inbox, which is as much
		// proof the address is theirs as a verification link would be.
		err := tx.Model(user).Updates(map[string]any{
			"password_hash":         user.PasswordHash,
			"is_temporary_password": false,
			"failed_login_count":    0,
			"locked_until":          nil,
			"email_verified":        true,
		}).Error
		if err != nil {
			return err
		}

		statements := []string{
			"UPDATE user_sessions SET revoked_at = @at WHERE user_id = @user AND revoked_at IS NULL",
			"UPDATE refresh_tokens SET revoked_at = @at WHERE user_id = @user AND revoked_at IS NULL",
			// Any other link sent to the same address stops working too.
			"UPDATE password_resets SET used_at = @at WHERE user_id = @user AND used_at IS NULL",
		}
		for _, statement := range statements {
			if err := tx.Exec(statement, map[string]any{"at": at, "user": user.ID}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// ---- A user's own account -------------------------------------------------

// ActiveUserSessions returns the sessions still signing a user in, newest
// first.
func (s *Store) ActiveUserSessions(ctx context.Context, user uuid.UUID, now time.Time) ([]model.UserSession, error) {
	var sessions []model.UserSession
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", user, now).
		Order("auth_time DESC").
		Find(&sessions).Error

	return sessions, err
}

// ChangeUserPassword saves a user's new password and, in the same transaction,
// ends every session but `keep` and revokes every refresh token they have.
func (s *Store) ChangeUserPassword(ctx context.Context, user *model.User, keep uuid.UUID, at time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(user).Updates(map[string]any{
			"password_hash":         user.PasswordHash,
			"is_temporary_password": false,
		}).Error
		if err != nil {
			return err
		}

		statements := []string{
			"UPDATE user_sessions SET revoked_at = @at WHERE user_id = @user AND id <> @keep AND revoked_at IS NULL",
			"UPDATE refresh_tokens SET revoked_at = @at WHERE user_id = @user AND revoked_at IS NULL",
		}
		for _, statement := range statements {
			if err := tx.Exec(statement, map[string]any{"at": at, "user": user.ID, "keep": keep}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Grant is what one application holds for a user: the scopes of its usable
// refresh tokens, when the user first signed in for it, and when it last got
// a token.
type Grant struct {
	Application model.Application
	Scopes      []string
	FirstIssued time.Time
	LastIssued  time.Time
}

// UserGrants lists the applications holding a usable refresh token for the
// user, most recently used first.
func (s *Store) UserGrants(ctx context.Context, user uuid.UUID, now time.Time) ([]Grant, error) {
	var tokens []model.RefreshToken
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", user, now).
		Order("created_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}

	byApp := map[uuid.UUID]*Grant{}
	order := []uuid.UUID{}
	for _, token := range tokens {
		grant, seen := byApp[token.ApplicationID]
		if !seen {
			grant = &Grant{FirstIssued: token.AuthTime, LastIssued: token.CreatedAt}
			byApp[token.ApplicationID] = grant
			order = append(order, token.ApplicationID)
		}

		grant.Scopes = append(grant.Scopes, token.Scope)
		if token.AuthTime.Before(grant.FirstIssued) {
			grant.FirstIssued = token.AuthTime
		}
	}

	out := make([]Grant, 0, len(order))
	for _, id := range order {
		app, err := s.Application(ctx, id)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}

		grant := byApp[id]
		grant.Application = *app
		out = append(out, *grant)
	}

	return out, nil
}

// RevokeUserGrant revokes every refresh token one application holds for a
// user, and says how many there were.
func (s *Store) RevokeUserGrant(ctx context.Context, user, application uuid.UUID, at time.Time) (int64, error) {
	result := s.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("user_id = ? AND application_id = ? AND revoked_at IS NULL", user, application).
		Update("revoked_at", at)

	return result.RowsAffected, result.Error
}

// ---- Signing key rotation -------------------------------------------------

// keyLock is the advisory lock that keeps two servers from rotating the
// signing keys at the same moment.
const keyLock = "xermess_signing_keys"

// AddSigningKeyIfDue stores `key` unless its algorithm already has an
// unretired key made after `dueBefore` — which another server may have just
// made. It holds an advisory lock while it looks, so of several servers
// deciding at once, one makes the key. It reports whether it did.
func (s *Store) AddSigningKeyIfDue(ctx context.Context, key *model.SigningKey, dueBefore time.Time) (bool, error) {
	added := false

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", keyLock).Error; err != nil {
			return err
		}

		var fresh int64
		err := tx.Model(&model.SigningKey{}).
			Where("algorithm = ? AND retired_at IS NULL AND created_at > ?", key.Algorithm, dueBefore).
			Count(&fresh).Error
		if err != nil || fresh > 0 {
			return err
		}

		added = true
		return translate(tx.Create(key).Error)
	})

	return added, err
}

// RetireSigningKeys retires every unretired key of an algorithm made before
// `before`: the ones a newer key has taken over from.
func (s *Store) RetireSigningKeys(ctx context.Context, algorithm string, before, at time.Time) error {
	return s.db.WithContext(ctx).Model(&model.SigningKey{}).
		Where("algorithm = ? AND retired_at IS NULL AND created_at < ?", algorithm, before).
		Update("retired_at", at).Error
}

// DeleteSigningKeys removes keys retired before `before`. Tokens they signed
// have expired, so they need not be published any more.
func (s *Store) DeleteSigningKeys(ctx context.Context, before time.Time) error {
	return s.db.WithContext(ctx).
		Where("retired_at IS NOT NULL AND retired_at < ?", before).
		Delete(&model.SigningKey{}).Error
}

// DeleteSigningKeysExcept removes every key but these: what revoking a
// compromised key means, since a published key is one APIs still trust.
func (s *Store) DeleteSigningKeysExcept(ctx context.Context, keep []string) error {
	// GORM names the KID field's column k_id.
	return s.db.WithContext(ctx).Where("k_id NOT IN ?", keep).Delete(&model.SigningKey{}).Error
}

// ---- Email verifications ----------------------------------------------------

// CreateEmailVerification stores a verification link that is about to be sent.
func (s *Store) CreateEmailVerification(ctx context.Context, verification *model.EmailVerification) error {
	return translate(s.db.WithContext(ctx).Omit(clause.Associations).Create(verification).Error)
}

// EmailVerificationByHash returns the link a token belongs to.
func (s *Store) EmailVerificationByHash(ctx context.Context, hash string) (*model.EmailVerification, error) {
	var verification model.EmailVerification
	if err := s.db.WithContext(ctx).First(&verification, "token_hash = ?", hash).Error; err != nil {
		return nil, translate(err)
	}

	return &verification, nil
}

// VerifyEmail uses a verification link: it marks the link, and every other
// one sent to the user, used, and the user's address verified, in one
// transaction. A link that carries a new address moves the account to it in
// the same transaction, so an address is never half changed.
//
// A link already used is ErrAlreadyUsed, and an address another account has
// taken since the link was sent is ErrDuplicate.
func (s *Store) VerifyEmail(ctx context.Context, verification *model.EmailVerification, at time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.EmailVerification{}).
			Where("id = ? AND used_at IS NULL", verification.ID).
			Update("used_at", at)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrAlreadyUsed
		}

		err := tx.Model(&model.EmailVerification{}).
			Where("user_id = ? AND used_at IS NULL", verification.UserID).
			Update("used_at", at).Error
		if err != nil {
			return err
		}

		changes := map[string]any{"email_verified": true}
		if verification.IsChange() {
			changes["email"] = model.NormalizeEmail(verification.NewEmail)
		}

		err = tx.Model(&model.User{}).Where("id = ?", verification.UserID).Updates(changes).Error

		return translate(err)
	})
}
