// Package validator provides request validation for the API.
package validator

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/pb33f/libopenapi"
	oapivalidator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi-validator/config"

	"github.com/rizesql/mithras/internal/errkit"
	"github.com/rizesql/mithras/pkg/api"
)

// ValidationErrors is a slice of api.FieldError that implements the error interface.
type ValidationErrors []api.FieldError

func (v ValidationErrors) Error() string {
	return "one or more fields failed validation"
}

// Validator wraps an oapivalidator.Validator and provides request validation for the API.
type Validator struct {
	validator oapivalidator.Validator
}

// New creates a new Validator instance.
func New() (*Validator, error) {
	document, err := libopenapi.NewDocument(api.Spec)
	if err != nil {
		return nil, errkit.Wrap(err, errkit.Internal("failed to create OpenAPI document"))
	}

	v, errors := oapivalidator.NewValidator(document, config.WithRegexCache(&sync.Map{}))
	if len(errors) > 0 {
		messages := make([]errkit.Option, len(errors))
		for i, e := range errors {
			messages[i] = errkit.Internal(e.Error())
		}

		return nil, errkit.New("failed to create validator", messages...)
	}

	if valid, docErrors := v.ValidateDocument(); !valid {
		messages := make([]errkit.Option, len(docErrors))
		for i, e := range docErrors {
			messages[i] = errkit.Internal(e.Message)
		}
		return nil, errkit.New("openapi document is invalid", messages...)
	}

	return &Validator{validator: v}, nil
}

// Validate checks the request and returns an errkit error if validation fails.
// If valid, it returns nil.
func (v *Validator) Validate(ctx context.Context, r *http.Request) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	valid, validationErrs := v.validator.ValidateHttpRequest(r)
	if valid {
		return nil
	}

	publicMsgs := make([]string, 0, len(validationErrs))
	fieldErrs := make(ValidationErrors, 0, len(validationErrs))

	var buf strings.Builder
	buf.Grow(80)

	for _, err := range validationErrs {
		for _, verr := range err.SchemaValidationErrors {
			buf.Reset()
			buf.WriteString("Field '")
			buf.WriteString(verr.FieldPath)
			buf.WriteString("': ")
			buf.WriteString(verr.Reason)
			publicMsgs = append(publicMsgs, buf.String())

			fieldErrs = append(fieldErrs, api.FieldError{
				Message: verr.Reason,
				Path:    verr.FieldPath,
				Hint:    &err.HowToFix,
			})
		}

		if len(err.SchemaValidationErrors) == 0 {
			publicMsgs = append(publicMsgs, fmt.Sprintf("%s: %s", err.ValidationType, err.Reason))
			fieldErrs = append(fieldErrs, api.FieldError{
				Message: err.Reason,
				Path:    err.ValidationType,
				Hint:    &err.HowToFix,
			})
		}
	}

	return errkit.Wrap(
		fieldErrs,
		errkit.Code(errkit.App.Validation.Code("openapi_mismatch")),
		errkit.Public("Validation failed: "+strings.Join(publicMsgs, " | ")),
	)
}
