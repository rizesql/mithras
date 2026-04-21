// Package errkit provides rich error wrapping with public/internal messages and error codes.
package errkit

import (
	"errors"
	"strings"
)

type richError struct {
	err      error
	public   string
	internal string
	code     Code
}

func (e *richError) Error() string { return e.err.Error() }
func (e *richError) Unwrap() error { return e.err }
func (e *richError) Is(target error) bool {
	if target == e {
		return true
	}

	return errors.Is(e.err, target)
}
func (e *richError) As(target any) bool {
	return errors.As(e.err, target)
}

// New creates a new rich error with the given message.
func New(msg string, opts ...Option) error {
	e := &richError{err: errors.New(msg)}
	for _, opt := range opts {
		opt.apply(e)
	}

	return e
}

// Wrap adds context to an existing error. Returns nil if err is nil.
func Wrap(err error, opts ...Option) error {
	if err == nil {
		return nil
	}

	if len(opts) == 0 {
		return err
	}

	e := &richError{err: err}
	for _, opt := range opts {
		opt.apply(e)
	}

	return e
}

// GetPublic collects all user-safe messages from the error chain.
func GetPublic(err error) string {
	if err == nil {
		return ""
	}

	var msgs []string

	for curr := range walk(err) {
		if re, ok := errors.AsType[*richError](curr); ok && re.public != "" {
			msgs = append(msgs, re.public)
		}
	}

	return strings.Join(msgs, " ")
}

// GetInternal collects all internal messages from the error chain.
func GetInternal(err error) string {
	if err == nil {
		return ""
	}

	var msgs []string

	for curr := range walk(err) {
		if re, ok := errors.AsType[*richError](curr); ok && re.internal != "" {
			msgs = append(msgs, re.internal)
		}
	}

	return strings.Join(msgs, ": ")
}

// GetCode returns the first non-zero error code from the chain.
func GetCode(err error) Code {
	for curr := range walk(err) {
		if re, ok := errors.AsType[*richError](curr); ok && !re.code.IsZero() {
			return re.code
		}
	}

	return Code{}
}
