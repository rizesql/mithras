package errkit

import (
	"errors"
	"iter"
)

func walk(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		walkImpl(err, yield)
	}
}

func walkImpl(err error, yield func(error) bool) bool {
	if err == nil {
		return true
	}

	if !yield(err) {
		return false
	}

	switch unwrapper := any(err).(type) {
	case interface{ Unwrap() []error }:
		for _, e := range unwrapper.Unwrap() {
			if !walkImpl(e, yield) {
				return false
			}
		}
	case interface{ Unwrap() error }:
		return walkImpl(unwrapper.Unwrap(), yield)
	}

	return true
}

// FlatError holds flattened error data from an error chain.
type FlatError struct {
	Public   []string
	Internal []string
	Code     Code
	Root     error
}

// Flatten extracts all error data from a chain into a single struct.
func Flatten(err error) FlatError {
	var f FlatError

	for curr := range walk(err) {
		re, ok := errors.AsType[*richError](curr)
		if !ok {
			if f.Root == nil {
				f.Root = curr
			}
			continue
		}

		if re.public != "" {
			f.Public = append(f.Public, re.public)
		}

		if re.internal != "" {
			f.Internal = append(f.Internal, re.internal)
		}

		if f.Code.IsZero() && !re.code.IsZero() {
			f.Code = re.code
		}
	}

	return f
}
