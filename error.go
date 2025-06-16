// error.go defines the core error struct used across the package.
// It encapsulates a message, optional stack trace, and an optional cause.
// Implements interfaces for unwrapping, formatting, and compatibility.

package errors

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// Error represents an error with optional stack trace and cause. Implements standard interfaces.
type Error struct {
	err             error     // wrapped error
	message         string    // contextual message for wrapping, not the final error string
	programCounters []uintptr // program counters for stack trace
	cachedStacks    []Stack   // cached stack traces to avoid recomputing
}

// Cause returns the underlying error.
func (e *Error) Cause() error {
	return e.Unwrap()
}

// Error returns the error message with wrapped messages joined by ": ".
func (e *Error) Error() string {
	if e.err == nil {
		return e.message
	}

	if e.message == "" {
		return e.err.Error()
	}

	if e.err.Error() == "" {
		return e.message
	}

	return fmt.Sprintf("%s: %s", e.message, e.err.Error())
}

// Format implements fmt.Formatter. Use %+v for full stack trace.
func (e *Error) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			_, _ = io.WriteString(state, e.Error())

			for _, stack := range e.Stacks() {
				_, _ = fmt.Fprintf(
					state,
					"\n%v\n\t%s:%d",
					stack.FuncName,
					stack.File,
					stack.LineNumber,
				)
			}

			return
		}

		fallthrough
	case 's':
		_, _ = io.WriteString(state, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(state, "%q", e.Error())
	}
}

// ProgramCounters returns the raw program counters (for testing).
func (e *Error) ProgramCounters() []uintptr {
	return e.programCounters
}

// Stacks returns the stack trace.
func (e *Error) Stacks() []Stack {
	if e.cachedStacks != nil {
		return e.cachedStacks
	}

	stacks := make([]Stack, 0, len(e.programCounters))

	for _, pc := range e.programCounters {
		stack := NewStack(pc)
		if strings.HasPrefix(stack.FuncName, "github.com/stackprune/errors.") {
			// Skip internal functions to avoid cluttering the stack trace
			continue
		}

		if slices.Contains(
			[]string{"runtime.goexit", "runtime.main", "testing.runExample"},
			stack.FuncName,
		) {
			break
		}

		stacks = append(stacks, stack)
	}

	e.cachedStacks = stacks

	return e.cachedStacks
}

// Unwrap returns the underlying cause error for compatibility with errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.err
}

var _ error = (*Error)(nil)
