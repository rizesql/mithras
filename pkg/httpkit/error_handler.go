package httpkit

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/api/validator"
	"github.com/rizesql/mithras/pkg/errkit"
	"github.com/rizesql/mithras/pkg/tracing"
)

// ProblemType represents a URI reference identifying the problem type
type ProblemType string

// Predefined problem types for common error scenarios
const (
	ProblemTypeValidationFailed  ProblemType = "https://api.mithras.com/problems/validation-failed"
	ProblemTypeUnauthorized      ProblemType = "https://api.mithras.com/problems/unauthorized"
	ProblemTypeForbidden         ProblemType = "https://api.mithras.com/problems/forbidden"
	ProblemTypeNotFound          ProblemType = "https://api.mithras.com/problems/not-found"
	ProblemTypeConflict          ProblemType = "https://api.mithras.com/problems/conflict"
	ProblemTypeRateLimitExceeded ProblemType = "https://api.mithras.com/problems/rate-limit-exceeded"
	ProblemTypeInternalError     ProblemType = "https://api.mithras.com/problems/internal-error"
	ProblemTypeBadGateway        ProblemType = "https://api.mithras.com/problems/bad-gateway"
	ProblemTypeTimeout           ProblemType = "https://api.mithras.com/problems/timeout"
	ProblemTypeUnknown           ProblemType = "https://api.mithras.com/problems/unknown"
)

// ErrorHandler is a function that handles errors returned by the server.
type ErrorHandler func(c *Context, err error)

func defaultErrorHandler(c *Context, err error) {
	msg := errkit.GetPublic(err)
	if msg == "" {
		msg = "An unexpected error occurred."
	}

	status := http.StatusInternalServerError
	code := errkit.GetCode(err)

	if !code.IsZero() {
		status = mapCodeToHTTPStatus(code)
	}

	problemType := mapCodeToProblemType(code)
	instance := c.req.raw.URL.Path

	if errs, ok := errors.AsType[validator.ValidationErrors](err); ok {
		response := api.BadRequestError{
			Type:      string(ProblemTypeValidationFailed),
			Title:     "Validation Failed",
			Status:    status,
			Detail:    new(fmt.Sprintf("%d field(s) failed validation", len(errs))),
			Errors:    errs,
			Instance:  new(instance),
			RequestId: c.Req().ID(),
		}

		if writeErr := c.res.ProblemJSON(status, response); writeErr != nil {
			tracing.Error("server.error_handler.write_json_failed",
				"error", writeErr)
		}
		return
	}

	response := api.Problem{
		Type:      string(problemType),
		Title:     http.StatusText(status),
		Status:    status,
		Detail:    new(msg),
		Instance:  new(instance),
		RequestId: c.Req().ID(),
	}

	if writeErr := c.res.ProblemJSON(status, response); writeErr != nil {
		tracing.Error("server.error_handler.write_json_failed",
			"error", writeErr)
	}
}

func mapCodeToHTTPStatus(code errkit.Code) int {
	scope := code.Scope()

	switch scope {
	case "user.request", "app.validation":
		return http.StatusBadRequest
	case "user.auth":
		return http.StatusUnauthorized
	case "user.forbidden", "user.permissions":
		return http.StatusForbidden
	case "app.resource.not_found":
		return http.StatusNotFound
	case "app.resource.conflict":
		return http.StatusConflict
	case "user.rate_limit":
		return http.StatusTooManyRequests

	case "app.internal", "app.db", "system":
		return http.StatusInternalServerError
	case "app.dependency":
		return http.StatusBadGateway
	case "system.timeout":
		return http.StatusGatewayTimeout
	}

	switch {
	case strings.HasPrefix(scope, "user."):
		return http.StatusBadRequest
	case strings.HasPrefix(scope, "app."):
		return http.StatusInternalServerError
	case strings.HasPrefix(scope, "system."):
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}

func mapCodeToProblemType(code errkit.Code) ProblemType {
	scope := code.Scope()

	switch scope {
	case "app.validation", "user.request":
		return ProblemTypeValidationFailed

	case "user.auth":
		return ProblemTypeUnauthorized

	case "user.forbidden", "user.permissions":
		return ProblemTypeForbidden

	case "app.resource.not_found":
		return ProblemTypeNotFound
	case "app.resource.conflict":
		return ProblemTypeConflict

	case "user.rate_limit":
		return ProblemTypeRateLimitExceeded

	case "app.internal", "app.db":
		return ProblemTypeInternalError
	case "app.dependency":
		return ProblemTypeBadGateway
	case "system.timeout":
		return ProblemTypeTimeout
	}

	switch {
	case strings.HasPrefix(scope, "app."):
		return ProblemTypeInternalError
	case strings.HasPrefix(scope, "user."):
		return ProblemTypeValidationFailed
	case strings.HasPrefix(scope, "system."):
		return ProblemTypeTimeout
	}

	return ProblemTypeUnknown
}
