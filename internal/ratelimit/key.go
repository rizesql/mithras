package ratelimit

import (
	"context"
	"fmt"
	"strings"

	"github.com/rizesql/mithras/pkg/httpkit"
)

// KeyFunc extracts a rate limit key from the incoming request. The returned
// string is used to namespace the counter in the Store. An empty string means
// the key could not be extracted; the middleware treats this as a no-op for
// that policy (the request is allowed through).
type KeyFunc func(ctx context.Context, c *httpkit.Context) string

// KeyIP keys on the client IP address resolved by httpkit. This delegates to
// c.Req().IP(), which already handles X-Forwarded-For parsing and port stripping,
// so there is no duplication of that logic here.
func KeyIP() KeyFunc {
	return func(_ context.Context, c *httpkit.Context) string {
		return c.Req().IP()
	}
}

// KeyBearerToken keys on the raw Bearer token from the Authorization header.
// Returns empty string if the header is absent or not a Bearer scheme.
func KeyBearerToken() KeyFunc {
	return func(_ context.Context, c *httpkit.Context) string {
		auth := c.Req().Raw().Header.Get("Authorization")
		if len(auth) <= 7 || !strings.EqualFold(auth[:7], "bearer ") {
			return ""
		}

		return auth[7:]
	}
}

// KeyBodyValue keys on a named field from the JSON request body. Because
// httpkit.Context.Init buffers the full body before middleware runs, the body
// can be unmarshalled here without consuming or interfering with downstream
// handlers.
//
// The value is extracted from a flat map[string]any — nested fields are not
// supported. Returns empty string if the field is absent or not a string.
func KeyBodyValue(field string, sanitizer ...func(string) string) KeyFunc {
	sanitizeFunc := noop[string]
	if len(sanitizer) > 0 {
		sanitizeFunc = sanitizer[0]
	}

	return func(_ context.Context, c *httpkit.Context) string {
		var body map[string]any
		if err := c.Req().BindBody(&body); err != nil {
			return ""
		}

		val, ok := body[field]
		if !ok || val == nil {
			return ""
		}

		return sanitizeFunc(fmt.Sprint(val))
	}
}

// KeyHeader keys on an arbitrary HTTP header value.
func KeyHeader(name string) KeyFunc {
	return func(_ context.Context, c *httpkit.Context) string {
		return sanitize(c.Req().Raw().Header.Get(name))
	}
}

func sanitize(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

func noop[T any](t T) T { return t }
