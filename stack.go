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
// It skips the first 3 frames to exclude internal function calls.
func callers() []uintptr {
	const (
		callersDepth = 32
		skipDepth    = 3
	)

	var pcs [callersDepth]uintptr
	length := runtime.Callers(skipDepth, pcs[:])

	return pcs[:length]
}
