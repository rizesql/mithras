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

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/telemetry"
	"github.com/rizesql/mithras/pkg/telemetry/logger"
)

// Dependencies holds the dependencies for the HTTP server.
type Dependencies struct {
	Clock        clock.Clock
	ErrorHandler ErrorHandler
}

// Server holds the state of the HTTP server.
type Server struct {
	pool       sync.Pool
	config     Config
	clock      clock.Clock
	errHandler ErrorHandler

	mu          sync.Mutex
	isListening bool

	mux *http.ServeMux
	srv *http.Server
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
	cx, ok := srv.pool.Get().(*Context)
	if !ok {
		panic("unable to acquire context from pool")
	}

	cx.reset()

	return cx
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

	withErrHandling := func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, c *Context) error {
			err := next(ctx, c)
			if err != nil {
				err = telemetry.Err(ctx, err)
				srv.errHandler(c, err)
			} else {
				telemetry.Ok(ctx)
			}

			return nil
		}
	}
	wrapped := withErrHandling(chain)

	handler := func(w http.ResponseWriter, r *http.Request) {
		cx := srv.acquire()

		defer func() {
			cx.reset()
			srv.release(cx)
		}()

		if err := cx.Init(w, r, srv.config.MaxRequestBodySize, readBody, srv.clock); err != nil {
			logger.Error("server.init_context_failed", "error", err)
			srv.errHandler(cx, err)

			return
		}

		wideCtx := telemetry.InjectMainSpan(r.Context(), trace.SpanFromContext(r.Context()))

		//nolint:errcheck // errors already handled by withErrHandling above
		_ = wrapped(withContext(wideCtx, cx), cx)
	}

	srv.mux.Handle(pattern, otelhttp.NewHandler(http.HandlerFunc(handler), pattern))
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
