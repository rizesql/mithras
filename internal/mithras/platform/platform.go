// Package platform provides the dependencies for Mithras.
package platform

import (
	"github.com/rizesql/mithras/internal/auth"
	"github.com/rizesql/mithras/internal/jws"
	"github.com/rizesql/mithras/internal/mithras/config"
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/internal/token"
	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
)

// Platform holds the dependencies for Mithras.
type Platform struct {
	Validator     *validator.Validator
	Clock         clock.Clock
	DB            *db.Database
	RateLimit     ratelimit.Store
	JWS           jws.Store
	Issuer        *token.Issuer
	Config        *config.Config
	OAuth2        *auth.OAuth2
	PasswordReset *auth.PasswordReset
}

// New creates a new Platform with the given dependencies.
func New(
	val *validator.Validator,
	clk clock.Clock,
	dbConn *db.Database,
	ra ratelimit.Store,
	jwsStore jws.Store,
	issuer *token.Issuer,
	cfg *config.Config,
	oa *auth.OAuth2,
	pr *auth.PasswordReset,
) *Platform {
	return &Platform{
		Validator:     val,
		Clock:         clk,
		DB:            dbConn,
		RateLimit:     ra,
		JWS:           jwsStore,
		Issuer:        issuer,
		Config:        cfg,
		OAuth2:        oa,
		PasswordReset: pr,
	}
}
