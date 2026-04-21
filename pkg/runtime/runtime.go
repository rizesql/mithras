// Package runtime provides application lifecycle management with graceful shutdown.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"

	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

// TaskFunc is a long-running task that respects context cancellation.
type TaskFunc func(ctx context.Context) error

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
	stateStopped
)

// Runtime manages application lifecycle with graceful shutdown support.
type Runtime struct {
	mu sync.Mutex
	wg sync.WaitGroup

	cleanups []ShutdownFunc
	state    state
	health   *healthState

	errCh  chan error
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Runtime instance.
func New(ctx context.Context) *Runtime {
	ctx, cancel := context.WithCancel(ctx)

	return &Runtime{
		health:   newHealthState(),
		cleanups: make([]ShutdownFunc, 0),
		errCh:    make(chan error, 1),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// fail triggers shutdown (first error wins)
func (rt *Runtime) fail(err error) {
	select {
	case rt.errCh <- err:
	default:
	}
	rt.cancel()
}

// Recover handles panics in goroutines and logs the stack trace.
func (rt *Runtime) Recover() {
	if r := recover(); r != nil {
		err := fmt.Errorf("panic: %v", r)
		logger.Error("runtime.panic",
			"panic", r,
			"stack", string(debug.Stack()),
		)
		rt.fail(err)
	}
}

// Go starts a background task managed by the runtime.
// Tasks are tracked and waited for during shutdown.
// Panics are recovered and logged. Errors are sent to the error channel.
func (rt *Runtime) Go(fn TaskFunc) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.state == stateShuttingDown {
		return
	}

	rt.wg.Go(func() {
		if err := fn(rt.ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("runtime.task_failed",
				"error", err,
			)
			rt.fail(err)
		}
	})
}

// Defer registers cleanup functions to run during shutdown.
// Functions run in reverse order of registration.
func (rt *Runtime) Defer(fns ...CloseFunc) {
	if len(fns) == 0 {
		return
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.state == stateShuttingDown {
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

	if rt.state == stateShuttingDown {
		return
	}

	rt.cleanups = append(rt.cleanups, fns...)
}

// // RunConfig holds runtime configuration options.
// type RunConfig struct {
// 	Signals          []os.Signal
// 	Timeout          time.Duration
// 	ReadinessTimeout time.Duration
// }

// // RunOption configures RunConfig.
// type RunOption func(*RunConfig)

// // WithTimeout sets the shutdown timeout.
// func WithTimeout(d time.Duration) RunOption {
// 	return func(c *RunConfig) {
// 		c.Timeout = d
// 	}
// }

// // WithSignals sets the OS signals to listen for.
// func WithSignals(sigs ...os.Signal) RunOption {
// 	return func(c *RunConfig) {
// 		c.Signals = sigs
// 	}
// }

// // WithReadinessTimeout sets the timeout for readiness checks.
// func WithReadinessTimeout(d time.Duration) RunOption {
// 	return func(c *RunConfig) {
// 		c.ReadinessTimeout = d
// 	}
// }

// // Run starts the runtime and blocks until shutdown.
// // It listens for signals, context cancellation, or task errors.
// // Returns an error if shutdown failed or a task errored.
// func (rt *Runtime) Run(ctx context.Context, opts ...RunOption) error {
// 	cfg := RunConfig{
// 		Timeout:          30 * time.Second,
// 		Signals:          []os.Signal{syscall.SIGINT, syscall.SIGTERM},
// 		ReadinessTimeout: 0,
// 	}
// 	for _, o := range opts {
// 		o(&cfg)
// 	}

// 	rt.mu.Lock()
// 	if rt.state.Load() == uint32(stateRunning) {
// 		rt.mu.Unlock()
// 		return errors.New("runtime is already running")
// 	}

// 	rt.state.Store(uint32(stateRunning))
// 	rt.health.mu.Lock()
// 	rt.health.checkTimeout = cfg.ReadinessTimeout
// 	rt.health.mu.Unlock()
// 	rt.mu.Unlock()

// 	logger.Info("runtime.started",
// 		"timeout", cfg.Timeout,
// 		"readiness_timeout", cfg.ReadinessTimeout,
// 	)

// 	sigCh := make(chan os.Signal, 1)

// 	signal.Notify(sigCh, cfg.Signals...)
// 	defer signal.Stop(sigCh)

// 	var err error

// 	reason := ""

// 	select {
// 	case sig := <-sigCh:
// 		reason = "signal: " + sig.String()
// 	case <-rt.ctx.Done():
// 		reason = "context canceled"
// 	case err = <-rt.errCh:
// 		reason = "task error"
// 	}

// 	if err != nil {
// 		logger.Error("runtime.shutdown_with_error", "error", err)
// 	} else {
// 		logger.Info("runtime.shutting_down", "reason", reason)
// 	}

// 	rt.mu.Lock()
// 	rt.state.Store(uint32(stateShuttingDown))
// 	cleanups := slices.Clone(rt.cleanups)
// 	rt.mu.Unlock()

// 	rt.cancel()

// 	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, cfg.Timeout)
// 	defer cancelShutdown()

// 	workersDone := make(chan struct{})

// 	go func() {
// 		rt.wg.Wait()
// 		close(workersDone)
// 	}()

// 	select {
// 	case <-workersDone:
// 		logger.Info("runtime.all_tasks_exited")
// 	case <-shutdownCtx.Done():
// 		logger.Error("runtime.shutdown_timeout")
// 	}

// 	var errs []error
// 	if err != nil {
// 		errs = append(errs, err)
// 	}

// 	for _, cl := range slices.Backward(cleanups) {
// 		if cl == nil {
// 			continue
// 		}

// 		if err := cl(shutdownCtx); err != nil {
// 			logger.Error("runtime.cleanup_failed",
// 				"error", err)
// 			errs = append(errs, err)
// 		}
// 	}

// 	if len(errs) > 0 {
// 		finalErr := errors.Join(errs...)
// 		logger.Error("runtime.shutdown_completed_with_errors", "error", finalErr)

// 		return finalErr
// 	}

// 	logger.Info("runtime.shutdown_complete")

// 	return nil
// }
