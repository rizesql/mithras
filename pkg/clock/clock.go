// Package clock provides time abstraction for testable code.
package clock

import (
	"sync"
	"time"
)

// Clock abstracts time retrieval for testability.
type Clock interface {
	Now() time.Time
}

// SystemClock returns the current system time.
type SystemClock struct{}

// Now returns the current system time.
func (c *SystemClock) Now() time.Time { return time.Now() }

// System is a global SystemClock instance.
var System Clock = &SystemClock{}

// TestClock is a fake clock for deterministic testing.
type TestClock struct {
	now time.Time
	mu  sync.RWMutex
}

// NewTestClock creates a TestClock initialized to the given time.
// If now is zero, uses the current time.
func NewTestClock(now time.Time) *TestClock {
	if now.IsZero() {
		now = time.Now()
	}

	return &TestClock{mu: sync.RWMutex{}, now: now}
}

var _ Clock = (*TestClock)(nil)

// Now returns the current fake time.
func (c *TestClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.now
}

// Tick advances the clock by d and returns the new time.
func (c *TestClock) Tick(d time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(d)

	return c.now
}

// Set sets the clock to t and returns t.
func (c *TestClock) Set(t time.Time) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = t

	return c.now
}
