// Package platform provides the dependencies for Mithras.
package platform

import (
	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/clock"
)

// Platform holds the dependencies for Mithras.
type Platform struct {
	Validator *validator.Validator
	Clock     clock.Clock
}

// New creates a new Platform with the given dependencies.
func New(v *validator.Validator, clk clock.Clock) *Platform {
	return &Platform{Validator: v, Clock: clk}
}
