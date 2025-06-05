// compat_pkg.go provides pkg/errors-compatible APIs not covered by the Go
// standard library, such as Errorf and Wrap. Unlike pkg/errors, this package
// avoids redundant stack traces and excludes WithMessage and similar clutter.

package errors

import (
	"fmt"
)

// WithStack adds a stack trace to the error, if not already present.
func WithStack(err error) error {
	return wrapWithMessage(err, "")
}

// Wrap wraps the given error with a message, preserving the original stack trace.
func Wrap(e error, message string) error {
	return wrapWithMessage(e, message)
}

// Wrapf wraps the error with a formatted message, similar to fmt.Sprintf.
func Wrapf(e error, format string, args ...any) error {
	return wrapWithMessage(e, fmt.Sprintf(format, args...))
}

// wrapWithMessage is a helper function that wraps an error with a message.
// It preserves stack trace information from existing Error types.
func wrapWithMessage(err error, message string) error {
	if err == nil {
		return nil
	}

	var (
		programCounters []uintptr
		cachedStacks    []Stack
	)

	var errorWithStack *Error
	if As(err, &errorWithStack) {
		programCounters = errorWithStack.programCounters
		cachedStacks = errorWithStack.cachedStacks
	} else {
		programCounters = callers()
	}

	return &Error{
		err:             err,
		message:         message,
		programCounters: programCounters,
		cachedStacks:    cachedStacks,
	}
}
