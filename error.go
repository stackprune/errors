// error.go defines the core error struct used across the package.
// It encapsulates a message, optional stack trace, and an optional cause.
// Implements interfaces for unwrapping, formatting, and compatibility.

package errors

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
)

// Error represents an error with optional stack trace and cause. Implements standard interfaces.
type Error struct {
	err             error       // wrapped error
	message         string      // contextual message for wrapping, not the final error string
	programCounters []uintptr   // program counters for stack trace
	cachedStacks    []Stack     // cached stack traces to avoid recomputing
	stackMutex      *sync.Mutex // guards cachedStacks without preventing Error value copies
}

// Cause returns the underlying error.
func (e *Error) Cause() error {
	return e.Unwrap()
}

// Error returns the error message with wrapped messages joined by ": ".
// The wrapped error's message is evaluated once per call.
func (e *Error) Error() string {
	if e.err == nil {
		return e.message
	}

	causeMessage := e.err.Error()

	if e.message == "" {
		return causeMessage
	}

	if causeMessage == "" {
		return e.message
	}

	return e.message + ": " + causeMessage
}

// Format implements fmt.Formatter. Use %+v for full stack trace.
func (e *Error) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			_, _ = io.WriteString(state, e.Error())

			e.printStack(state)

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
// It is safe to call concurrently.
func (e *Error) Stacks() []Stack {
	if e.stackMutex == nil {
		return []Stack{}
	}

	e.stackMutex.Lock()
	defer e.stackMutex.Unlock()

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

// cachedStackFrames returns the current cache without computing it.
func (e *Error) cachedStackFrames() []Stack {
	if e.stackMutex == nil {
		return nil
	}

	e.stackMutex.Lock()
	defer e.stackMutex.Unlock()

	return e.cachedStacks
}

// printStack prints the stack trace without the error message.
func (e *Error) printStack(state fmt.State) {
	for _, stack := range e.Stacks() {
		_, _ = fmt.Fprintf(
			state,
			"\n%v\n\t%s:%d",
			stack.FuncName,
			stack.File,
			stack.LineNumber,
		)
	}
}

var _ error = (*Error)(nil)
