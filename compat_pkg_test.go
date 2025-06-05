package errors_test

import (
	stderrors "errors"
	"io"
	"testing"

	"github.com/stackprune/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithStackNil(t *testing.T) {
	t.Parallel()

	got := errors.WithStack(nil)
	assert.NoError(t, got)
}

func TestWithStack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input error
		want  string
	}{
		{
			name:  "error input",
			input: io.EOF,
			want:  "EOF",
		},
		{
			name:  "standard error input",
			input: stderrors.New("standard error"),
			want:  "standard error",
		},
		{
			name:  "stackprune error input",
			input: errors.New("wrapped error"),
			want:  "wrapped error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.WithStack(tt.input).Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   error
		message string
		want    string
	}{
		{
			name:    "nil input",
			input:   nil,
			message: "additional context",
			want:    "",
		},
		{
			name:    "error input",
			input:   stderrors.New("original error"),
			message: "database failed",
			want:    "database failed: original error",
		},
		{
			name:    "stackprune error input",
			input:   errors.New("wrapped error"),
			message: "service failed",
			want:    "service failed: wrapped error",
		},
		{
			name:    "multiple wraps",
			input:   errors.Wrap(errors.New("root"), "level1"),
			message: "level2",
			want:    "level2: level1: root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Wrap(tt.input, tt.message)

			if tt.input == nil {
				assert.NoError(t, got)

				return
			}

			require.Error(t, got)
			require.Equal(t, tt.want, got.Error())

			// Check that stack trace is preserved/added
			stackErr := &errors.Error{}
			ok := stderrors.As(got, &stackErr)
			assert.True(t, ok)
			assert.NotEmpty(t, stackErr.Stacks())
		})
	}
}

func TestWrapf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  error
		format string
		args   []any
		want   string
	}{
		{
			name:   "nil input",
			input:  nil,
			format: "failed with code %d",
			args:   []any{500},
			want:   "",
		},
		{
			name:   "simple format",
			input:  stderrors.New("connection failed"),
			format: "database error",
			args:   nil,
			want:   "database error: connection failed",
		},
		{
			name:   "format with args",
			input:  stderrors.New("timeout"),
			format: "operation failed after %d seconds",
			args:   []any{30},
			want:   "operation failed after 30 seconds: timeout",
		},
		{
			name:   "multiple format args",
			input:  stderrors.New("permission denied"),
			format: "user %s cannot access %s",
			args:   []any{"john", "/admin"},
			want:   "user john cannot access /admin: permission denied",
		},
		{
			name:   "stackprune error input",
			input:  errors.New("original"),
			format: "wrapped with %s",
			args:   []any{"context"},
			want:   "wrapped with context: original",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Wrapf(tt.input, tt.format, tt.args...)

			if tt.input == nil {
				assert.NoError(t, got)

				return
			}

			require.Error(t, got)
			require.Equal(t, tt.want, got.Error())

			// Check that stack trace is preserved/added
			stackErr := &errors.Error{}
			ok := stderrors.As(got, &stackErr)
			require.True(t, ok)
			assert.NotEmpty(t, stackErr.Stacks())
		})
	}
}
