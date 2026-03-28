package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rizesql/mithras/pkg/tracing"
)

const (
	defaultCheckTimeout = 500 * time.Millisecond
)

var (
	respNotStarted   = []byte(`{"status":"not started"}`)
	respShuttingDown = []byte(`{"status":"shutting down"}`)
	respOK           = []byte(`{"status":"ok"}`)
)

// ReadinessCheck is a health check function for readiness probes.
type ReadinessCheck func(ctx context.Context) error

// healthState holds the state of health checks.
type healthState struct {
	mu           sync.RWMutex
	checks       map[string]ReadinessCheck
	checkTimeout time.Duration
}

// newHealthState creates a new healthState with default timeout.
func newHealthState() *healthState {
	return &healthState{
		checks:       make(map[string]ReadinessCheck),
		checkTimeout: defaultCheckTimeout,
	}
}

// AddReadiness registers a readiness check with the given name.
// Panics if name is empty, check is nil, or name is already registered.
func (rt *Runtime) AddReadiness(name string, check ReadinessCheck) {
	if name == "" {
		panic("runtime: readiness check name cannot be empty")
	}
	if check == nil {
		panic("runtime: readiness check function cannot be nil")
	}

	rt.health.mu.Lock()
	defer rt.health.mu.Unlock()

	if _, exists := rt.health.checks[name]; exists {
		panic("runtime: readiness check '" + name + "' is already registered")
	}

	rt.health.checks[name] = check
}

// RegisterHealth registers /health/live and /health/ready endpoints on the mux.
// The prefix parameter optionally sets a base path (default: "/health").
func (rt *Runtime) RegisterHealth(mux *http.ServeMux, prefix ...string) {
	base := "/health"
	if len(prefix) > 0 {
		base = strings.TrimRight(prefix[0], "/")
	}

	tracing.Info("runtime.health_registered",
		"base", base)

	mux.HandleFunc(fmt.Sprintf("GET %s/live", base), rt.handleLive)
	mux.HandleFunc(fmt.Sprintf("GET %s/ready", base), rt.handleReady)
}

// handleLive returns the liveness probe response.
// Returns 503 if runtime hasn't started, 200 otherwise.
func (rt *Runtime) handleLive(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if rt.state.Load() == uint32(stateIdle) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write(respNotStarted)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respOK)
}

// handleReady returns the readiness probe response.
// Returns 503 if not started, shutting down, or any check fails.
func (rt *Runtime) handleReady(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	currentState := rt.state.Load()
	if currentState == uint32(stateIdle) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write(respNotStarted)
		return
	}
	if currentState == uint32(stateShuttingDown) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write(respShuttingDown)
		return
	}

	rt.health.mu.RLock()
	checks := make(map[string]ReadinessCheck, len(rt.health.checks))
	maps.Copy(checks, rt.health.checks)
	rt.health.mu.RUnlock()

	if len(checks) == 0 {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(respOK)
		return
	}

	results := rt.runChecks(req.Context(), checks)

	response := struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}{
		Status: "ok",
		Checks: make(map[string]string, len(results)),
	}

	allPassed := true
	for _, result := range results {
		if result.err != nil {
			response.Checks[result.name] = result.err.Error()
			allPassed = false
		} else {
			response.Checks[result.name] = "ok"
		}
	}

	if !allPassed {
		response.Status = "fail"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(response)
}

type checkResult struct {
	name string
	err  error
}

// runChecks executes all health checks concurrently with a timeout.
// Returns results for all checks, with timeout errors for slow checks.
func (rt *Runtime) runChecks(ctx context.Context, checks map[string]ReadinessCheck) []checkResult {
	results := make([]checkResult, 0, len(checks))
	resultCh := make(chan checkResult, len(checks))

	batchCtx, cancelBatch := context.WithTimeout(ctx, rt.health.checkTimeout)
	defer cancelBatch()

	for name, check := range checks {
		go func(name string, check ReadinessCheck) {
			defer func() {
				if r := recover(); r != nil {
					resultCh <- checkResult{name: name, err: fmt.Errorf("panic: %v", r)}
				}
			}()

			err := check(batchCtx)
			resultCh <- checkResult{name: name, err: err}
		}(name, check)
	}

	for range len(checks) {
		select {
		case res := <-resultCh:
			results = append(results, res)
		case <-batchCtx.Done():
			finished := make(map[string]bool)
			for _, r := range results {
				finished[r.name] = true
			}

			for name := range checks {
				if !finished[name] {
					results = append(results, checkResult{
						name: name,
						err:  fmt.Errorf("health check timed out after %v", rt.health.checkTimeout),
					})
				}
			}
			return results
		}
	}

	return results
}
