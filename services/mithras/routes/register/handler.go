package register

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/rizesql/mithras/internal/email"
	"github.com/rizesql/mithras/internal/errkit"
	"github.com/rizesql/mithras/internal/password"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/db"
	"github.com/rizesql/mithras/pkg/httpkit"
	"github.com/rizesql/mithras/pkg/httpkit/middleware"
	"github.com/rizesql/mithras/pkg/idkit"
	"github.com/rizesql/mithras/pkg/telemetry"
	"github.com/rizesql/mithras/services/mithras/platform"
)

type (
	Req api.V2RegisterRequest
	Res api.V2RegisterResponse
)

type handler struct {
	db *db.Database
}

func New(p *platform.Platform) *handler {
	return &handler{db: p.DB}
}

func (h *handler) Method() string { return http.MethodPost }
func (h *handler) Path() string   { return "/v2/register" }

func (h *handler) Handle(ctx context.Context, c *httpkit.Context) error {
	req, err := httpkit.BindBody[Req](c)
	if err != nil {
		return err
	}

	eAddr, err := email.Parse(string(req.Email))
	if err != nil {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.User.Request.Code("invalid_email")),
			errkit.Internal("email validation failed"),
			errkit.Publicf("%s", err.Error()),
		)
	}

	telemetry.Attr(ctx, attribute.String("user.email", eAddr.String()))

	pwd, err := password.New(req.Password)
	if err != nil {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.User.Request.Code("invalid_password")),
			errkit.Internal("password validation failed"),
			errkit.Publicf("%s", err.Error()),
		)
	}

	secret, err := hashPassword(ctx, pwd)
	if err != nil {
		return errkit.Wrap(err,
			errkit.WithCode(errkit.App.Internal.Code("password_hash_failed")),
			errkit.Internal("failed to hash password"),
			errkit.Public("Failed to register user."),
		)
	}

	userID := idkit.NewUserID()
	telemetry.Attr(ctx, attribute.String("user.id", string(userID)))

	err = db.TxRetry(ctx, h.db, func(tx db.DBTX) error {
		if err := db.Query.InsertUser(ctx, tx, db.InsertUserParams{
			ID:    userID,
			Name:  req.Name,
			Email: eAddr,
		}); err != nil {
			if db.IsDuplicateError(err) {
				telemetry.Attr(ctx, attribute.Bool("registration.duplicate_email", true))
				return errkit.New("user with email already exists",
					errkit.WithCode(errkit.User.Request.Code("duplicate_email")),
					errkit.Publicf("A user with email %v already exists", string(req.Email)),
					errkit.Internalf("duplicate user"),
				)
			}

			return errkit.Wrap(err,
				errkit.WithCode(errkit.System.Code("service_unavailable")),
				errkit.Internal("database error"),
				errkit.Public("Failed to register user."),
			)
		}

		if err := db.Query.InsertCredential(ctx, tx, db.InsertCredentialParams{
			UserID: userID,
			Secret: *secret,
		}); err != nil {
			return errkit.Wrap(err,
				errkit.WithCode(errkit.System.Code("service_unavailable")),
				errkit.Internal("failed to insert credential"),
				errkit.Public("Failed to register user."),
			)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return c.Res().JSON(http.StatusCreated, Res{
		Id: string(userID),
	})
}

func RateLimit(p *platform.Platform) httpkit.Middleware {
	return middleware.WithRateLimit(
		ratelimit.NewPolicy("register-per-ip",
			10, time.Minute,
			ratelimit.KeyIP(),
			ratelimit.WithStore(p.RateLimit),
			ratelimit.WithBurst(),
		),

		ratelimit.NewPolicy("register-per-account",
			5, time.Minute,
			ratelimit.KeyBodyValue("email", strings.ToLower),
			ratelimit.WithStore(p.RateLimit),
		),
	)
}

func hashPassword(ctx context.Context, pwd password.Raw) (*password.Hashed, error) {
	spanCtx, span := telemetry.Start(ctx, "hash_password")
	defer span.End()

	secret, err := pwd.Hash()
	if err != nil {
		telemetry.Err(spanCtx, err)
		return nil, err
	}

	return &secret, nil
}
