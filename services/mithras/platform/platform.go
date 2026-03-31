// Package platform provides the dependencies for Mithras.
package platform

import (
	"github.com/rizesql/mithras/internal/ratelimit"
	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/db"
)

// Platform holds the dependencies for Mithras.
type Platform struct {
	Validator *validator.Validator
	Clock     clock.Clock
	DB        *db.Database
	RateLimit ratelimit.Store
}

// New creates a new Platform with the given dependencies.
func New(v *validator.Validator, clk clock.Clock, d *db.Database, ra ratelimit.Store) *Platform {
	return &Platform{Validator: v, Clock: clk, DB: d, RateLimit: ra}
}
