package httpkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rizesql/mithras/internal/errkit"
	"github.com/rizesql/mithras/pkg/clock"
	"github.com/rizesql/mithras/pkg/idkit"
)

// Request holds the incoming HTTP request details.
type Request struct {
	id        string
	ip        string
	timestamp time.Time
	raw       *http.Request
	body      []byte
}

// ID returns the unique request ID.
func (req *Request) ID() string { return req.id }

// IP returns the client's IP address.
func (req *Request) IP() string { return req.ip }

// Timestamp returns the time the request was received.
func (req *Request) Timestamp() time.Time { return req.timestamp }

// Raw returns the underlying *http.Request.
func (req *Request) Raw() *http.Request { return req.raw }

// BindBody unmarshals the request body into the given destination.
func (req *Request) BindBody(dst any) error {
	if err := json.Unmarshal(req.body, dst); err != nil {
		return errkit.Wrap(err,
			errkit.Code(errkit.User.Request.Code("invalid_json_body")),
			errkit.Internal("failed to unmarshal request body"),
			errkit.Public("The request body was not valid JSON."),
		)
	}

	return nil
}

// Response holds the outgoing HTTP response details.
type Response struct {
	w    statusRecorder
	body []byte
}

// StatusCode returns the HTTP status code of the response.
func (res *Response) StatusCode() int { return res.w.statusCode }

// AddHeader adds a header to the response.
func (res *Response) AddHeader(key, val string) {
	res.w.Header().Add(key, val)
}

// SetHeader sets a header in the response.
func (res *Response) SetHeader(key, val string) {
	res.w.Header().Set(key, val)
}

func (res *Response) send(status int, body []byte) error {
	res.body = body

	res.w.WriteHeader(status)
	if _, err := res.w.Write(body); err != nil {
		return errkit.Wrap(err,
			errkit.Code(errkit.App.Internal.Code("response_write_failed")),
			errkit.Internal("failed to send bytes"),
			errkit.Public("Unable to send response body."),
		)
	}

	return nil
}

// Send sends a raw response with the given status code and body.
func (res *Response) Send(status int, body []byte) error {
	return res.send(status, body)
}

// JSON sends a JSON response with the given status code and body.
func (res *Response) JSON(status int, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return errkit.Wrap(
			err,
			errkit.Code(errkit.App.Internal.Code("response_marshal_failed")),
			errkit.Internal("json marshal failed"),
			errkit.Public("The response body could not be marshalled to JSON."),
		)
	}

	res.w.Header().Add("Content-Type", "application/json")
	return res.send(status, b)
}

// ProblemJSON sends a JSON response with the given status code and body, using the "application/problem+json" content type.
func (res *Response) ProblemJSON(status int, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return errkit.Wrap(
			err,
			errkit.Code(errkit.App.Internal.Code("response_marshal_failed")),
			errkit.Internal("json marshal failed"),
			errkit.Public("The response body could not be marshalled to JSON."),
		)
	}

	res.w.Header().Add("Content-Type", "application/problem+json")
	return res.send(status, b)
}

// Context holds the request and response context for an HTTP request.
type Context struct {
	req Request
	res Response
}

// Req returns a pointer to the request context.
func (c *Context) Req() *Request { return &c.req }

// Res returns a pointer to the response context.
func (c *Context) Res() *Response { return &c.res }

// Init initializes the request and response contexts for an HTTP request.
func (c *Context) Init(w http.ResponseWriter, r *http.Request, maxBodySize int64, readBody bool, clk clock.Clock) error {
	c.req.id = idkit.New("req")
	c.req.ip = resolveIP(r)
	c.req.timestamp = clk.Now()
	c.req.raw = r

	c.res.w = statusRecorder{ResponseWriter: w}
	c.Res().AddHeader("X-Request-Id", c.req.id)

	if !readBody {
		_ = r.Body.Close()
		return nil
	}

	if maxBodySize > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	}

	var err error
	c.req.body, err = io.ReadAll(r.Body)
	closeErr := r.Body.Close()

	if err != nil {
		if maxBytesErr, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return errkit.Wrap(err,
				errkit.Code(errkit.User.Request.Code("request_body_too_large")),
				errkit.Internal(fmt.Sprintf("request body exceeds size limit of %d bytes", maxBytesErr.Limit)),
				errkit.Public(fmt.Sprintf("The request body exceeds the maximum allowed size of %d bytes.", maxBytesErr.Limit)),
			)
		}

		return errkit.Wrap(err,
			errkit.Code(errkit.User.Request.Code("body_read_failed")),
			errkit.Internal("unable to read request body"),
			errkit.Public("The request body could not be read."),
		)
	}

	if closeErr != nil {
		return errkit.Wrap(closeErr,
			errkit.Code(errkit.App.Internal.Code("body_close_failed")),
			errkit.Internal("failed to close request body"),
			errkit.Public("An error occurred processing the request."),
		)
	}

	r.Body = io.NopCloser(bytes.NewReader(c.req.body))
	return nil
}

func (c *Context) reset() {
	const maxRetainedCapacity = 1 << 20 // 1 MB

	c.req.id = ""
	c.req.ip = ""
	c.req.timestamp = time.Time{}
	c.req.raw = nil
	if cap(c.req.body) > maxRetainedCapacity {
		c.req.body = nil
	} else {
		c.req.body = c.req.body[:0]
	}

	c.res.w = statusRecorder{}
	if cap(c.res.body) > maxRetainedCapacity {
		c.res.body = nil
	} else {
		c.res.body = c.res.body[:0]
	}
}

func resolveIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for ip := range strings.SplitSeq(xff, ",") {
			if ip = strings.TrimSpace(ip); ip != "" {
				return stripPort(ip)
			}
		}
	}

	return stripPort(r.RemoteAddr)
}

func stripPort(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}

	return addr
}

// BindBody binds the request body to a struct of type T.
func BindBody[T any](c *Context) (T, error) {
	var req T

	if err := c.Req().BindBody(&req); err != nil {
		return req, errkit.Wrap(err,
			errkit.Code(errkit.App.Validation.Code("invalid_input")),
			errkit.Internal("invalid request body"),
			errkit.Public("The request body is invalid."),
		)
	}

	return req, nil
}
