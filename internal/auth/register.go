package auth

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rizesql/mithras/internal/email"
	"github.com/rizesql/mithras/internal/password"
	"github.com/rizesql/mithras/internal/token"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/idkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

type Register struct {
	db    *db.Database
	login *Login
}

func NewRegister(d *db.Database, clk clock.Clock, iss *token.Issuer, cfg *Config) Register {
	return Register{db: d, login: NewLogin(d, clk, iss, cfg)}
}

type RegisterResult struct {
	UserID  idkit.UserID
	UserPk  int64
	Session *LoginResponse
}

func (r Register) Register(
	ctx context.Context,
	name, rawEmail, rawPassword, userAgent, ipAddr string,
) (res *RegisterResult, err error) {
	ctx, span := telemetry.Start(ctx, "auth.Register")
	defer telemetry.End(span, &err)

	addr, err := email.Parse(rawEmail)
	if err != nil {
		return nil, err
	}

	pwd, err := password.New(rawPassword)
	if err != nil {
		return nil, err
	}

	secret, err := hashPassword(ctx, pwd)
	if err != nil {
		return nil, errPasswordHashFailed(err)
	}

	type txResult struct {
		userPk int64
		userID idkit.UserID
	}

	result, err := db.TxWithResultRetry(ctx, r.db, func(tx db.DBTX) (txResult, error) {
		userID := idkit.NewUserID()

		userPk, err := db.Query.InsertUser(ctx, tx, db.InsertUserParams{
			ID:    userID,
			Name:  name,
			Email: addr,
		})
		if err != nil {
			if db.IsDuplicateError(err) {
				telemetry.Attr(ctx, attribute.Bool("registration.duplicate_email", true))

				return txResult{}, errDuplicateEmail(addr.Raw())
			}

			return txResult{}, errRegistrationDatabaseError(err)
		}

		if err = db.Query.InsertCredential(ctx, tx, db.InsertCredentialParams{
			UserPk: userPk,
			Secret: *secret,
		}); err != nil {
			return txResult{}, errCredentialInsertFailed(err)
		}

		if err = db.Query.InsertPasswordHistory(ctx, tx, db.InsertPasswordHistoryParams{
			UserPk: userPk,
			Secret: *secret,
		}); err != nil {
			return txResult{}, errPasswordHistoryInsertFailed(err)
		}

		if err = db.Query.AssignRole(ctx, tx, db.AssignRoleParams{
			UserPk:    userPk,
			Name:      "USER",
			GrantedBy: pgtype.Int8{Int64: 0, Valid: false},
		}); err != nil {
			return txResult{}, errDefaultRoleAssignmentFailed(err)
		}

		return txResult{userPk: userPk, userID: userID}, nil
	})
	if err != nil {
		return nil, err
	}

	sess, err := r.login.CreateSession(ctx, result.userPk, result.userID, userAgent, ipAddr)
	if err != nil {
		return &RegisterResult{UserID: result.userID}, nil
	}

	return &RegisterResult{
		UserID:  result.userID,
		UserPk:  result.userPk,
		Session: sess,
	}, nil
}

func hashPassword(ctx context.Context, pwd password.Raw) (secret *password.Hashed, err error) {
	_, span := telemetry.Start(ctx, "auth.hash_password")
	defer telemetry.End(span, &err)

	h, err := pwd.Hash()
	if err != nil {
		return nil, err
	}

	return &h, nil
}
