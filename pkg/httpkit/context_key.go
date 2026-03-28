package httpkit

import "context"

type contextKey[T any] struct {
	name string
}

func newContextKey[T any](name string) contextKey[T] {
	return contextKey[T]{name: name}
}

func (k contextKey[T]) withValue(ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, k, value)
}

var ctxKey = newContextKey[*Context]("httpkit.context")

func withContext(ctx context.Context, s *Context) context.Context {
	return ctxKey.withValue(ctx, s)
}
