// Package runtime provides application lifecycle management with graceful shutdown.
package runtime

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime/debug"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/rizesql/mithras/pkg/tracing"
)

// RunFunc is a long-running task that respects context cancellation.
type RunFunc func(ctx context.Context) error

// ShutdownFunc is called during graceful shutdown.
type ShutdownFunc func(ctx context.Context) error

// CloseFunc is a shutdown function that doesn't need context.
type CloseFunc func() error

// state represents the runtime lifecycle state.
type state uint32

// Runtime lifecycle states.
const (
	stateIdle state = iota
	stateRunning
	stateShuttingDown
)

// Runtime manages application lifecycle with graceful shutdown support.
type Runtime struct {
	mu sync.Mutex
	wg sync.WaitGroup

	cleanups []ShutdownFunc

	state  atomic.Uint32
	health *healthState
	errCh  chan error

	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Runtime instance.
func New(ctx context.Context) *Runtime {
	//nolint:gosec // cancel function is stored in struct and called during Shutdown
	// #nosec G118
	ctx, cancel := context.WithCancel(ctx)
	return &Runtime{
		health:   newHealthState(),
		cleanups: make([]ShutdownFunc, 0),
		errCh:    make(chan error, 1),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Recover handles panics in goroutines and logs the stack trace.
func (rt *Runtime) Recover() {
	if recovered := recover(); recovered != nil {
		tracing.Error("panic",
			"panic", recovered,
			"stack", string(debug.Stack()),
		)
	}
}

// Go starts a background task managed by the runtime.
// Tasks are tracked and waited for during shutdown.
// Panics are recovered and logged. Errors are sent to the error channel.
func (rt *Runtime) Go(fn RunFunc) {
	rt.mu.Lock()
	if rt.state.Load() == uint32(stateShuttingDown) {
		rt.mu.Unlock()
		return
	}
	rt.wg.Add(1)
	rt.mu.Unlock()

	go func() {
		defer rt.wg.Done()
		defer rt.Recover()

		if err := fn(rt.ctx); err != nil && !errors.Is(err, context.Canceled) {
			tracing.Error("runtime.task_failed",
				"error", err)

			select {
			case rt.errCh <- err:
			default:
			}
		}
	}()
}

// Defer registers cleanup functions to run during shutdown.
// Functions run in reverse order of registration.
func (rt *Runtime) Defer(fns ...CloseFunc) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.state.Load() == uint32(stateShuttingDown) {
		return
	}

	for _, fn := range fns {
		rt.cleanups = append(rt.cleanups, func(context.Context) error {
			return fn()
		})
	}
}

// DeferFunc registers context-aware cleanup functions to run during shutdown.
// Functions run in reverse order of registration.
func (rt *Runtime) DeferFunc(fns ...ShutdownFunc) {
	if len(fns) == 0 {
		return
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.state.Load() == uint32(stateShuttingDown) {
		return
	}
	rt.cleanups = append(rt.cleanups, fns...)
}

// RunConfig holds runtime configuration options.
type RunConfig struct {
	Timeout          time.Duration
	Signals          []os.Signal
	ReadinessTimeout time.Duration
}

// RunOption configures RunConfig.
type RunOption func(*RunConfig)

// WithTimeout sets the shutdown timeout.
func WithTimeout(d time.Duration) RunOption {
	return func(c *RunConfig) {
		c.Timeout = d
	}
}

// WithSignals sets the OS signals to listen for.
func WithSignals(sigs ...os.Signal) RunOption {
	return func(c *RunConfig) {
		c.Signals = sigs
	}
}

// WithReadinessTimeout sets the timeout for readiness checks.
func WithReadinessTimeout(d time.Duration) RunOption {
	return func(c *RunConfig) {
		c.ReadinessTimeout = d
	}
}

// Run starts the runtime and blocks until shutdown.
// It listens for signals, context cancellation, or task errors.
// Returns an error if shutdown failed or a task errored.
func (rt *Runtime) Run(opts ...RunOption) error {
	cfg := RunConfig{
		Timeout:          30 * time.Second,
		Signals:          []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		ReadinessTimeout: 0,
	}
	for _, o := range opts {
		o(&cfg)
	}

	rt.mu.Lock()
	if rt.state.Load() == uint32(stateRunning) {
		rt.mu.Unlock()
		return errors.New("runtime is already running")
	}
	rt.state.Store(uint32(stateRunning))
	rt.health.mu.Lock()
	rt.health.checkTimeout = cfg.ReadinessTimeout
	rt.health.mu.Unlock()
	rt.mu.Unlock()

	tracing.Info("runtime.started",
		"timeout", cfg.Timeout,
		"readiness_timeout", cfg.ReadinessTimeout,
	)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, cfg.Signals...)
	defer signal.Stop(sigCh)

	var err error
	reason := ""
	select {
	case sig := <-sigCh:
		reason = "signal: " + sig.String()
	case <-rt.ctx.Done():
		reason = "context canceled"
	case err = <-rt.errCh:
		reason = "task error"
	}

	if err != nil {
		tracing.Error("runtime.shutdown_with_error", "error", err)
	} else {
		tracing.Info("runtime.shutting_down", "reason", reason)
	}

	rt.mu.Lock()
	rt.state.Store(uint32(stateShuttingDown))
	cleanups := slices.Clone(rt.cleanups)
	rt.mu.Unlock()

	rt.cancel()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancelShutdown()

	workersDone := make(chan struct{})
	go func() {
		rt.wg.Wait()
		close(workersDone)
	}()

	select {
	case <-workersDone:
		tracing.Info("runtime.all_tasks_exited")
	case <-shutdownCtx.Done():
		tracing.Error("runtime.shutdown_timeout")
	}

	var errs []error
	if err != nil {
		errs = append(errs, err)
	}

	for _, cl := range slices.Backward(cleanups) {
		if cl == nil {
			continue
		}
		if err := cl(shutdownCtx); err != nil {
			tracing.Error("runtime.cleanup_failed",
				"error", err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		finalErr := errors.Join(errs...)
		tracing.Error("runtime.shutdown_completed_with_errors", "error", finalErr)
		return finalErr
	}

	tracing.Info("runtime.shutdown_complete")
	return nil
}
