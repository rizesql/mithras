package auth

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/internal/email"
	"github.com/rizesql/mithras/internal/password"
	"github.com/rizesql/mithras/internal/token"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/idkit"
	"github.com/rizesql/mithras/pkg/telemetry"
)

type Login struct {
	db  *db.Database
	clk clock.Clock
	iss *token.Issuer
	cfg *Config
}

func NewLogin(d *db.Database, clk clock.Clock, iss *token.Issuer, cfg *Config) *Login {
	return &Login{
		db:  d,
		clk: clk,
		iss: iss,
		cfg: cfg,
	}
}

type LoginResponse struct {
	AccessToken  token.Access
	RefreshToken token.Refresh
}

func (l Login) Authenticate(
	ctx context.Context,
	rawEmail, rawPassword string,
) (usr db.GetUserWithPasswordRow, err error) {
	ctx, span := telemetry.Start(ctx, "auth.Authenticate")
	defer telemetry.End(span, &err)

	addr, pwd, err := parseCredentials(rawEmail, rawPassword)
	if err != nil {
		return db.GetUserWithPasswordRow{}, err
	}

	usr, err = l.fetchUser(ctx, addr)
	userNotFound := errors.Is(err, errInvalidCredentials)
	if err != nil && !userNotFound {
		return db.GetUserWithPasswordRow{}, err
	}

	now := l.clk.Now()
	statusErr := checkUserStatus(usr.Status, usr.LockedUntil, now)

	ok, verifyErr := verifyPassword(ctx, usr.Secret, pwd)

	if userNotFound {
		return db.GetUserWithPasswordRow{}, errInvalidCredentials
	}

	if statusErr != nil {
		return db.GetUserWithPasswordRow{}, statusErr
	}

	if verifyErr != nil {
		return db.GetUserWithPasswordRow{}, errPasswordVerificationFailed(verifyErr)
	}

	if !ok {
		return db.GetUserWithPasswordRow{}, l.handleLoginFailure(ctx, usr, now)
	}

	return usr, nil
}

func (l Login) fetchUser(
	ctx context.Context,
	addr email.Address,
) (db.GetUserWithPasswordRow, error) {
	usr, err := db.Query.GetUserWithPassword(ctx, l.db, addr)
	if err != nil {
		if db.IsNotFound(err) {
			return db.GetUserWithPasswordRow{}, errInvalidCredentials
		}
		return db.GetUserWithPasswordRow{}, errSessionLookupFailed(err)
	}

	return usr, nil
}

func (l Login) handleLoginFailure(
	ctx context.Context,
	usr db.GetUserWithPasswordRow,
	now time.Time,
) error {
	if err := db.Query.RecordLoginFailure(ctx, l.db, usr.Pk); err != nil {
		telemetry.Event(ctx, "auth.record_login_failure_failed",
			attribute.String("error", err.Error()),
		)
	}

	if int(usr.FailedAttempts)+1 >= l.cfg.MaxFailedAttempts {
		lockedUntil := now.Add(l.cfg.LockoutDuration)
		if err := db.Query.LockAccount(ctx, l.db, db.LockAccountParams{
			UserPk:      usr.Pk,
			LockedUntil: &lockedUntil,
		}); err != nil {
			telemetry.Event(ctx, "auth.lock_account_failed",
				attribute.String("error", err.Error()),
			)
		}
	}

	return errInvalidCredentials
}

func (l Login) CreateSession(
	ctx context.Context,
	userPk int64,
	userID idkit.UserID,
	userAgent, ipAddr string,
) (resp *LoginResponse, err error) {
	ctx, span := telemetry.Start(ctx, "auth.CreateSession")
	defer telemetry.End(span, &err)

	roles, err := db.Query.GetUserRoles(ctx, l.db, userPk)
	if err != nil {
		return nil, errRolesLookupFailed(err)
	}

	now := l.clk.Now()

	accessToken, err := l.iss.Issue(ctx, token.IssueConfig{
		Subject:  userID.String(),
		IssuedAt: now,
		Duration: l.cfg.AccessTokenDuration,
		Roles:    roles,
	})
	if err != nil {
		return nil, errTokenSigningFailed(err)
	}

	sess, err := newSession(now, l.cfg.RefreshTokenDuration)
	if err != nil {
		return nil, errTokenGenerationFailed(err)
	}

	telemetry.Attr(ctx, attribute.Bool("http.has_user_agent", userAgent != ""))

	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	err = db.Query.InsertSession(ctx, l.db, db.InsertSessionParams{
		ID:        sess.ID,
		UserPk:    userPk,
		TokenHash: sess.Token.Hash(),
		UserAgent: userAgentPtr,
		IpAddr:    netip.MustParseAddr(ipAddr),
		ExpiresAt: sess.ExpiresAt,
	})
	if err != nil {
		return nil, errSessionInsertFailed(err)
	}

	telemetry.Attr(ctx,
		attribute.Bool("auth.success", true),
		attribute.String("user.id", userID.String()),
		attribute.String("session.id", sess.ID.String()),
	)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: sess.Token,
	}, nil
}

func (l Login) Login(
	ctx context.Context,
	rawEmail, rawPassword, userAgent, ipAddr string,
) (resp *LoginResponse, err error) {
	ctx, span := telemetry.Start(ctx, "auth.Login")
	defer telemetry.End(span, &err)

	usr, err := l.Authenticate(ctx, rawEmail, rawPassword)
	if err != nil {
		telemetry.Attr(ctx,
			attribute.Bool("auth.success", false),
			attribute.String("auth.failure_reason", "invalid_credentials"),
		)
		return nil, err
	}

	if err := db.Query.RecordLoginSuccess(ctx, l.db, usr.Pk); err != nil {
		telemetry.Event(ctx, "auth.record_login_success_failed",
			attribute.String("error", err.Error()),
		)
	}

	return l.CreateSession(ctx, usr.Pk, usr.ID, userAgent, ipAddr)
}

func verifyPassword(ctx context.Context, secret password.Hashed, pwd password.Raw) (bool, error) {
	_, verifySpan := telemetry.Start(ctx, "auth.verify_password")
	defer telemetry.End(verifySpan, nil)

	return secret.Verify(pwd)
}
