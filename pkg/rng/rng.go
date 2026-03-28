// Package rng provides simple, testable random number abstractions.
package rng

import "math/rand"

// Rand produces float64 values in [0.0, 1.0).
type Rand interface {
	Float64() float64
}

// MathRand wraps math/rand for fast, non-cryptographic randomness.
type MathRand struct {
	r *rand.Rand
}

// NewMathRand returns a MathRand seeded with the given value.
func NewMathRand(seed int64) *MathRand {
	//nolint:gosec // MathRand is used while knowing it's non-cryptographic randomness
	// #nosec G404
	return &MathRand{r: rand.New(rand.NewSource(seed))}
}

// Float64 returns a pseudo-random value in [0.0, 1.0).
func (m *MathRand) Float64() float64 { return m.r.Float64() }

// DeterministicRand always returns the same value.
type DeterministicRand struct {
	Value float64
}

// Float64 returns the configured value.
func (d DeterministicRand) Float64() float64 { return d.Value }
