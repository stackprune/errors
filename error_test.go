package errors_test

import (
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stackprune/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("base error")

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "simple error",
			err:  baseErr,
			want: "base error",
		},
		{
			name: "wrapped error",
			err:  errors.Wrap(baseErr, "wrapped"),
			want: "wrapped: base error",
		},
		{
			name: "multiple wraps",
			err:  errors.Wrap(errors.Wrap(baseErr, "level1"), "level2"),
			want: "level2: level1: base error",
		},
		{
			name: "wrapped empty string error",
			err:  errors.Wrap(errors.New(""), "wrapped"),
			want: "wrapped",
		},
		{
			name: "base error with empty message",
			err:  errors.Wrap(baseErr, ""),
			want: "base error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.err.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestError_Format(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("test error")
	wrappedErr := wrapError()

	tests := []struct {
		name       string
		err        error
		format     string
		want       string
		wantRegexp *regexp.Regexp
	}{
		{
			name:   "simple string format",
			err:    baseErr,
			format: "%s",
			want:   "test error",
		},
		{
			name:   "quoted format",
			err:    baseErr,
			format: "%q",
			want:   `"test error"`,
		},
		{
			name:   "simple verbose format",
			err:    baseErr,
			format: "%v",
			want:   "test error",
		},
		{
			name:   "detailed verbose format",
			err:    baseErr,
			format: "%+v",
			wantRegexp: regexp.MustCompile(
				`(?s)^test error\n` +
					`github\.com/stackprune/errors_test\.TestError_Format\n` +
					`\t[\S]+/error_test\.go:\d+\n`,
			),
		},
		{
			name:   "wrapped error simple string format",
			err:    wrappedErr,
			format: "%s",
			want:   "wrapped: test error",
		},
		{
			name:   "wrapped error quoted format",
			err:    wrappedErr,
			format: "%q",
			want:   `"wrapped: test error"`,
		},
		{
			name:   "wrapped error detailed verbose format",
			err:    wrappedErr,
			format: "%+v",
			wantRegexp: regexp.MustCompile(
				`(?s)^wrapped: test error\n` +
					`github\.com/stackprune/errors_test\.wrapError\n` +
					`\t[\S]+/testutil_test\.go:11\n` +
					`github\.com/stackprune/errors_test\.TestError_Format\n` +
					`\t[\S]+/error_test\.go:\d+\n`,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := fmt.Sprintf(tt.format, tt.err)
			if tt.wantRegexp == nil {
				assert.Equal(t, tt.want, got)
			} else {
				assert.Regexp(t, tt.wantRegexp, got)
			}
		})
	}
}

func TestError_Stacks(t *testing.T) {
	t.Parallel()

	var err *errors.Error

	require.ErrorAs(t, wrapError(), &err)
	stacks := err.Stacks()

	// Check that we have stacks
	require.NotEmpty(t, stacks)

	// Check that stacks are cached
	require.Equal(t, stacks, err.Stacks())

	wants := []struct {
		baseFileName string
		funcName     string
	}{
		{
			baseFileName: "testutil_test.go",
			funcName:     "github.com/stackprune/errors_test.wrapError",
		},
		{
			baseFileName: "error_test.go",
			funcName:     "github.com/stackprune/errors_test.TestError_Stacks",
		},
	}

	for index, want := range wants {
		got := stacks[index]

		assert.Equal(t, want.baseFileName, filepath.Base(got.File))
		assert.Equal(t, want.funcName, got.FuncName)
		assert.GreaterOrEqual(t, got.LineNumber, 1)
		assert.GreaterOrEqual(t, got.ProgramCounter, uintptr(1))
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("base error")
	wrappedErr := errors.Wrap(baseErr, "wrapped error")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "wrapped error",
			err:  wrappedErr,
			want: baseErr,
		},
		{
			name: "simple error",
			err:  baseErr,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stackpruneErr := func() *errors.Error {
				var target *errors.Error
				_ = errors.As(tt.err, &target)

				return target
			}()

			got := stackpruneErr.Unwrap()
			assert.Equal(t, tt.want, got)

			// Cause() is alias for Unwrap()
			got = stackpruneErr.Cause()
			assert.Equal(t, tt.want, got)
		})
	}
}
