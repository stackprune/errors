// core.go provides the primary constructors for error values such as
// New, Wrap, Errorf, and WithStack. These functions form the public
// API surface, and are fully compatible with the Go 1.13+ standard errors package.

package errors

import (
	"errors"
	"fmt"
)

// New creates an error with a message.
// It initializes an Error struct with the provided message.
func New(message string) error {
	return &Error{
		err:             nil,
		message:         message,
		programCounters: callers(),
		cachedStacks:    nil,
	}
}

// As reports whether any error in err’s chain matches target, same as errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Errorf formats a message and returns an error with a stack trace.
func Errorf(format string, args ...any) error {
	return &Error{
		err:             nil,
		message:         fmt.Sprintf(format, args...),
		programCounters: callers(),
		cachedStacks:    nil,
	}
}

// Is reports whether err is or wraps target, same as errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if any.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
