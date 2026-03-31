package errkit

import "fmt"

// Option configures a rich error.
type Option interface {
	apply(*richError)
}

type internalOpt struct{ msg string }

func (o internalOpt) apply(e *richError) { e.internal = o.msg }

// Internal adds an internal debug message to the error.
func Internal(msg string) Option {
	return internalOpt{msg: msg}
}

// Internalf adds a formatted internal debug message to the error.
func Internalf(format string, args ...any) Option {
	return Internal(fmt.Sprintf(format, args...))
}

type publicOpt struct{ msg string }

func (o publicOpt) apply(e *richError) { e.public = o.msg }

// Public adds a user-safe message to the error.
func Public(msg string) Option {
	return publicOpt{msg: msg}
}

// Publicf adds a formatted user-safe message to the error.
func Publicf(format string, args ...any) Option {
	return Public(fmt.Sprintf(format, args...))
}

type codeOpt struct{ code Code }

func (o codeOpt) apply(e *richError) { e.code = o.code }

// WithCode attaches an error code for classification.
func WithCode(c Code) Option {
	return codeOpt{code: c}
}
