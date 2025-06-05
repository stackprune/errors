// join_error.go defines the joinedError struct, which represents a group
// of errors returned by errors.Join. It implements methods to support errors.Is,
// errors.As, and custom formatting for composed errors.

package errors

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// JoinError represents a composed error returned by Join. Supports Is and As via Unwrap.
type JoinError struct {
	errs []error
}

// Error returns the concatenated error messages from all errors.
// Multiple errors are joined with newlines.
func (j *JoinError) Error() string {
	parts := make([]string, 0, len(j.errs))

	for _, err := range j.errs {
		if err != nil {
			parts = append(parts, err.Error())
		}
	}

	return strings.Join(parts, "\n")
}

// Format implements fmt.Formatter for JoinError.
// Use %+v to display stack traces for the first Error type only.
func (j *JoinError) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			j.printVerbose(state)

			return
		}

		fallthrough
	case 's':
		_, _ = io.WriteString(state, j.Error())
	case 'q':
		_, _ = fmt.Fprintf(state, "%q", j.Error())
	}
}

// Unwrap returns the slice of errors for compatibility with errors.Is and errors.As.
func (j *JoinError) Unwrap() []error {
	return j.errs
}

// printVerbose prints detailed error information with stack traces.
// Shows stack trace only for the first Error type to avoid repetition.
func (j *JoinError) printVerbose(state fmt.State) {
	printedStack := false

	for i, err := range j.errs {
		if i > 0 {
			_, _ = io.WriteString(state, "\n")
		}

		var errorWithStack *Error
		if errors.As(err, &errorWithStack) {
			if !printedStack {
				_, _ = fmt.Fprintf(state, "%+v", errorWithStack) // with stack

				printedStack = true
			} else {
				_, _ = fmt.Fprintf(state, "%s", errorWithStack) // without stack
			}
		} else {
			_, _ = fmt.Fprintf(state, "%+v", err)
		}
	}
}
