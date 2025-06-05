package errors_test

import (
	"fmt"

	"github.com/stackprune/errors"
)

// wrapError creates a wrapped error for testing purposes.
func wrapError() error {
	err := errors.New("test error")

	return errors.Wrap(err, "wrapped")
}

// customError is a test error type used across multiple test files.
type customError struct {
	message string
}

func (e customError) Error() string {
	return fmt.Sprintf("customError(%s)", e.message)
}

// newCustomError creates a new customError with the given message.
func newCustomError(message string) customError {
	return customError{message: message}
}
