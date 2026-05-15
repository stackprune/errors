// stack.go implements stack trace capturing via runtime.Callers and defines
// types for storing and formatting stack trace information.
// This logic is invoked internally by New and WithStack.

package errors

import (
	"runtime"
)

// Stack captures a single stack frame, including file, line, and function.
type Stack struct {
	File           string
	LineNumber     int
	FuncName       string
	ProgramCounter uintptr
}

// NewStack constructs a Stack from a raw program counter using runtime metadata.
func NewStack(programCounter uintptr) Stack {
	//nolint:exhaustruct
	stack := Stack{ProgramCounter: programCounter}
	if stack.ProgramCounter == 0 {
		return stack
	}

	pcFunc := runtime.FuncForPC(stack.ProgramCounter - 1)
	if pcFunc == nil {
		return stack
	}

	stack.FuncName = pcFunc.Name()
	stack.File, stack.LineNumber = pcFunc.FileLine(programCounter - 1)

	return stack
}

// callers captures the current stack trace.
// It skips the first 3 frames plus additional frames to exclude internal function calls.
// The skip parameter allows callers to skip additional frames as needed.
func callers(skip int) []uintptr {
	const (
		callersDepth     = 32
		defaultSkipDepth = 3
	)

	var pcs [callersDepth]uintptr

	length := runtime.Callers(defaultSkipDepth+skip, pcs[:])

	return pcs[:length]
}
