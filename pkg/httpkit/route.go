package httpkit

import "context"

// Handler is an interface for handling HTTP requests.
type Handler interface {
	Handle(ctx context.Context, c *Context) error
}

// HandleFunc is a function type for handling HTTP requests.
type HandleFunc func(ctx context.Context, c *Context) error

// Middleware is a function type for wrapping HandleFuncs.
type Middleware func(handler HandleFunc) HandleFunc

// Route is an interface for defining HTTP routes.
type Route interface {
	Handler

	Method() string
	Path() string
}
