package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/rizesql/mithras/internal/email"
	"github.com/rizesql/mithras/internal/password"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/idkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

type PasswordReset struct {
	db  *db.Database
	clk clock.Clock
}

func NewPasswordReset(d *db.Database, clk clock.Clock) *PasswordReset {
	return &PasswordReset{
		db:  d,
		clk: clk,
	}
}

func (r PasswordReset) Request(
	ctx context.Context,
	rawEmail string,
	userAgent, ipAddr string,
) (err error) {
	ctx, span := telemetry.Start(ctx, "auth.PasswordReset.Request")
	defer telemetry.End(span, &err)

	addr, err := email.Parse(rawEmail)
	if err != nil {
		return nil
	}

	usr, err := db.Query.GetUserWithPassword(ctx, r.db, addr)
	if err != nil {
		if db.IsNotFound(err) {
			return nil
		}
		return errUserLookupFailed(err)
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return errResetTokenGenerationFailed(err)
	}

	secretStr := base64.RawURLEncoding.EncodeToString(secret)
	hash := sha256.Sum256([]byte(secretStr))
	id := idkit.NewPasswordResetID()

	now := r.clk.Now()
	expiresAt := now.Add(time.Hour)

	_, err = db.Query.PasswordResetInsert(ctx, r.db, db.PasswordResetInsertParams{
		ID:        id,
		UserPk:    usr.Pk,
		TokenHash: hash[:],
		UserAgent: &userAgent,
		IpAddr:    netip.MustParseAddr(ipAddr),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return errTokenInsertFailed(err)
	}

	// FOR TESTING: Log the reset token so it can be used without SMTP
	//nolint:lll
	fmt.Printf("\n--- [DEBUG] PASSWORD RESET TOKEN ---\nToken: %s.%s\n------------------------------------\n\n", id, secretStr)

	// TODO: Send email
	// link := fmt.Sprintf("https://mithras.app/reset-password?token=%s.%s", id, secretStr)
	// r.email.SendResetLink(addr, link)

	return nil
}

func (r PasswordReset) Reset(
	ctx context.Context,
	token, rawNewPassword string,
) (err error) {
	ctx, span := telemetry.Start(ctx, "auth.PasswordReset.Reset")
	defer telemetry.End(span, &err)

	rst, err := r.validateToken(ctx, token)
	if err != nil {
		return err
	}

	newSecret, err := r.validateAndHashNewPassword(ctx, rst.UserPk, rawNewPassword)
	if err != nil {
		return err
	}

	return r.performReset(ctx, rst, newSecret)
}

func (r PasswordReset) validateToken(
	ctx context.Context,
	token string,
) (db.PasswordResetGetActiveRow, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return db.PasswordResetGetActiveRow{}, errInvalidResetToken
	}

	id := parts[0]
	secretStr := parts[1]

	rst, err := db.Query.PasswordResetGetActive(ctx, r.db, idkit.PasswordResetID(id))
	if err != nil {
		if db.IsNotFound(err) {
			return db.PasswordResetGetActiveRow{}, errResetTokenNotFound
		}
		return db.PasswordResetGetActiveRow{}, errTokenLookupFailed(err)
	}

	hash := sha256.Sum256([]byte(secretStr))
	if subtle.ConstantTimeCompare(hash[:], rst.TokenHash) != 1 {
		return db.PasswordResetGetActiveRow{}, errResetTokenSecretMismatch
	}

	return rst, nil
}

func (r PasswordReset) validateAndHashNewPassword(
	ctx context.Context,
	userPk int64,
	rawNewPassword string,
) (*password.Hashed, error) {
	pwd, err := password.New(rawNewPassword)
	if err != nil {
		return nil, err
	}

	newSecret, err := hashPassword(ctx, pwd)
	if err != nil {
		return nil, err
	}

	if err := r.checkPasswordHistory(ctx, userPk, pwd); err != nil {
		return nil, err
	}

	return newSecret, nil
}

func (r PasswordReset) checkPasswordHistory(
	ctx context.Context,
	userPk int64,
	pwd password.Raw,
) error {
	history, err := db.Query.GetRecentPasswordHashes(ctx, r.db, db.GetRecentPasswordHashesParams{
		UserPk: userPk,
		Limit:  5,
	})
	if err != nil {
		return errPasswordHistoryLookupFailed(err)
	}

	for _, oldSecret := range history {
		match, err := oldSecret.Verify(pwd)
		if err != nil {
			return errPasswordHistoryVerificationFailed(err)
		}
		if match {
			return errPasswordRecentlyUsed
		}
	}

	return nil
}

func (r PasswordReset) performReset(
	ctx context.Context,
	rst db.PasswordResetGetActiveRow,
	newSecret *password.Hashed,
) error {
	now := r.clk.Now()

	err := db.Tx(ctx, r.db, func(tx db.DBTX) error {
		if err := db.Query.InsertPasswordHistory(ctx, tx, db.InsertPasswordHistoryParams{
			UserPk: rst.UserPk,
			Secret: *newSecret,
		}); err != nil {
			return err
		}

		if err := db.Query.UpdateCredentialByUserId(ctx, tx, db.UpdateCredentialByUserIdParams{
			UserPk: rst.UserPk,
			Secret: *newSecret,
		}); err != nil {
			return err
		}

		if err := db.Query.PasswordResetMarkUsed(ctx, tx, rst.Pk); err != nil {
			return err
		}

		err := db.Query.PasswordResetInvalidateSiblings(ctx, tx, db.PasswordResetInvalidateSiblingsParams{
			UserPk: rst.UserPk,
			Pk:     rst.Pk,
		})
		if err != nil {
			return err
		}

		if err := db.Query.RevokeUserSessions(ctx, tx, db.RevokeUserSessionsParams{
			UserPk: rst.UserPk,
			Now:    &now,
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return errPasswordResetTransactionFailed(err)
	}

	return nil
}
