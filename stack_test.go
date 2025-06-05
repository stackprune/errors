package errors_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stackprune/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewStack(t *testing.T) {
	t.Parallel()

	isInvalidPC := func(pc uintptr) func(errors.Stack) bool {
		return func(stack errors.Stack) bool {
			return stack.ProgramCounter == pc &&
				stack.File == "" &&
				stack.LineNumber == 0 &&
				stack.FuncName == ""
		}
	}

	tests := []struct {
		name string
		pc   uintptr
		want func(errors.Stack) bool
	}{
		{
			name: "zero program counter",
			pc:   0,
			want: isInvalidPC(0),
		},
		{
			name: "invalid program counter",
			pc:   999999,
			want: isInvalidPC(999999),
		},
		{
			name: "valid program counter",
			pc:   getCurrentPC(),
			want: func(stack errors.Stack) bool {
				return stack.ProgramCounter != 0 &&
					strings.Contains(stack.File, "stack_test.go") &&
					stack.LineNumber > 0 &&
					strings.Contains(stack.FuncName, "getCurrentPC")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stack := errors.NewStack(tt.pc)
			assert.True(t, tt.want(stack), "Stack: %+v", stack)
		})
	}
}

func TestStackIntegration(t *testing.T) {
	t.Parallel()

	// Create an error and check its stacks
	err := func() *errors.Error {
		target := &errors.Error{}
		_ = errors.As(errors.New("test error"), &target)

		return target
	}()
	stacks := err.Stacks()

	assert.NotEmpty(t, stacks)

	// Look for any test-related function in the stack
	found := false

	for _, stack := range stacks {
		if strings.Contains(stack.FuncName, "Test") || strings.Contains(stack.File, "_test.go") {
			assert.Positive(t, stack.LineNumber)
			assert.NotEqual(t, uintptr(0), stack.ProgramCounter)

			found = true

			break
		}
	}

	assert.True(t, found, "Test-related function not found in stack trace")
}

func TestStackWithWrap(t *testing.T) {
	t.Parallel()

	// Test that wrapped errors preserve the original stack
	originalErr := errors.New("original error")
	wrappedErr := errors.Wrap(originalErr, "wrapped error")

	originalStacks := func() *errors.Error {
		target := &errors.Error{}
		_ = errors.As(originalErr, &target)

		return target
	}().Stacks()
	wrappedStacks := func() *errors.Error {
		target := &errors.Error{}
		_ = errors.As(wrappedErr, &target)

		return target
	}().Stacks()

	assert.Equal(t, originalStacks, wrappedStacks)
}

// getCurrentPC is a helper function to get the current program counter.
func getCurrentPC() uintptr {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		return 0
	}

	return pc
}
