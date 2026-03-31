// Package httpkit provides a simple HTTP server implementation.
package httpkit

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/logger"
)

// Dependencies holds the dependencies for the HTTP server.
type Dependencies struct {
	Clock        clock.Clock
	ErrorHandler ErrorHandler
}

// Server holds the state of the HTTP server.
type Server struct {
	mu          sync.Mutex
	isListening bool

	mux *http.ServeMux
	srv *http.Server

	config     Config
	errHandler ErrorHandler
	clock      clock.Clock

	pool sync.Pool
}

// New creates a new Server instance.
func New(deps Dependencies, cfg Config) *Server {
	if deps.ErrorHandler == nil {
		deps.ErrorHandler = defaultErrorHandler
	}

	mux := http.NewServeMux()

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{
		mu:          sync.Mutex{},
		isListening: false,
		mux:         mux,
		srv:         srv,
		config:      cfg,
		errHandler:  deps.ErrorHandler,
		clock:       deps.Clock,
		pool: sync.Pool{
			New: func() any {
				return &Context{
					req: Request{body: []byte{}},
					res: Response{body: []byte{}},
				}
			},
		},
	}
}

func (srv *Server) acquire() *Context {
	c, ok := srv.pool.Get().(*Context)
	if !ok {
		panic("unable to acquire context from pool")
	}
	c.reset()
	return c
}

func (srv *Server) release(c *Context) { srv.pool.Put(c) }

// Mux returns the underlying http.ServeMux of the server.
func (srv *Server) Mux() *http.ServeMux { return srv.mux }

// Serve starts the HTTP server and listens for incoming requests.
func (srv *Server) Serve(ctx context.Context, ln net.Listener) error {
	srv.mu.Lock()
	if srv.isListening {
		logger.Warn("server.already_listening")
		srv.mu.Unlock()
		return nil
	}
	srv.isListening = true
	srv.mu.Unlock()

	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			logger.Info("server.shutdown_requested")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			defer shutdownCancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Error("server.shutdown_failed",
					"error", err.Error())
			}
		case <-serverCtx.Done():
		}
	}()

	logger.Info("server.listening",
		"srv", "http",
		"addr", ln.Addr().String())
	err := srv.srv.Serve(ln)

	cancel()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server.listen_failed", "error", err)
		return err
	}
	return nil
}

// RegisterRoute registers a route with the server.
func (srv *Server) RegisterRoute(route Route, middlewares ...Middleware) {
	path := route.Path()
	method := route.Method()
	logger.Info("server.register_route",
		"method", method,
		"path", path)

	pattern := path
	if method != "" {
		pattern = method + " " + path
	}

	chain := HandleFunc(route.Handle)
	for _, mw := range slices.Backward(middlewares) {
		chain = mw(chain)
	}

	readBody := methodHasBody(method)

	// Wrap the chain so that error handling runs inside the middleware stack.
	// This ensures that when middleware defers fire (e.g. to capture status code),
	// the error handler has already written the actual HTTP status.
	withErrHandling := func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, c *Context) error {
			err := next(ctx, c)
			if err != nil {
				srv.errHandler(c, err)
			}
			return nil
		}
	}
	wrapped := withErrHandling(chain)

	srv.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		c := srv.acquire()
		defer func() {
			c.reset()
			srv.release(c)
		}()

		if err := c.Init(w, r, srv.config.MaxRequestBodySize, readBody, srv.clock); err != nil {
			logger.Error("server.init_context_failed", "error", err)
			srv.errHandler(c, err)
			return
		}

		//nolint:errcheck // errors already handled by withErrHandling above
		_ = wrapped(withContext(r.Context(), c), c)
	})
}

// Shutdown gracefully shuts down the server.
func (srv *Server) Shutdown(ctx context.Context) error {
	srv.mu.Lock()
	srv.isListening = false
	srv.mu.Unlock()

	if err := srv.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}

func methodHasBody(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete,
		http.MethodOptions, http.MethodTrace, http.MethodConnect:
		return false
	default:
		return true
	}
}
