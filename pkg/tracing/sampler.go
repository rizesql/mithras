package tracing

import (
	"time"

	"github.com/rizesql/mithras/pkg/rng"
)

// EventSnapshot holds a snapshot of an event for sampling purposes.
type EventSnapshot struct {
	ErrorCount int
	Duration   time.Duration
}

// Sampler determines whether an event should be emitted when [End] is called.
type Sampler interface {
	Sample(evt EventSnapshot) bool
}

// AlwaysSample is a sampler that always returns true, emitting all events. Not
// recommended for production use.
type AlwaysSample struct{}

// Sample always returns true, emitting all events.
func (s AlwaysSample) Sample(EventSnapshot) bool { return true }

// TailSampler is a probabilistic sampler with bias towards errors and slow requests.
type TailSampler struct {
	rng rng.Rand

	// LatencyThreshold is the duration beyond which events are sampled.
	LatencyThreshold time.Duration

	// SampleRate is the probability of sampling an event that isn't exceeding
	// [LatencyThreshold] or has errors.
	SampleRate float64
}

// Sample returns true if the event should be sampled, based on the configured rates.
func (s TailSampler) Sample(evt EventSnapshot) bool {
	if evt.ErrorCount > 0 {
		return true
	}

	if evt.Duration > s.LatencyThreshold {
		return true
	}

	rate := s.rng.Float64()
	return rate < s.SampleRate
}
