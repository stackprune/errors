package errors_test

import (
	"fmt"
	"testing"

	"github.com/stackprune/errors"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		want    error
	}{
		{
			name:    "empty message",
			message: "",
			want:    errors.New(""),
		},
		{
			name:    "simple message",
			message: "foo",
			want:    errors.New("foo"),
		},
		{
			name:    "message with format specifiers",
			message: "with format specifiers: %v",
			want:    errors.New("with format specifiers: %v"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.New(tt.message).Error()

			assert.Equal(t, tt.want.Error(), got)
		})
	}
}

func TestAs(t *testing.T) {
	t.Parallel()

	testErr := newCustomError("some error")

	tests := []struct {
		name   string
		err    error
		target any
		want   bool
	}{
		{
			name:   "nil error",
			err:    nil,
			target: &testErr,
			want:   false,
		},
		{
			name:   "with stack",
			err:    errors.WithStack(testErr),
			target: &testErr,
			want:   true,
		},
		{
			name:   "with wrap",
			err:    errors.Wrap(testErr, "wrapped error"),
			target: &testErr,
			want:   true,
		},
		{
			name:   "standard error compatibility",
			err:    fmt.Errorf("wrap with:%w", testErr),
			target: &testErr,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var target any
			switch v := tt.target.(type) {
			case *customError:
				target = &customError{}
			default:
				target = v
			}

			got := errors.As(tt.err, target)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		args    []any
		want    string
	}{
		{
			name:    "simple message",
			message: "foo",
			want:    "foo",
		},
		{
			name:    "message with format specifiers",
			message: "with format specifier, %s %d",
			args:    []any{"foo", 1},
			want:    "with format specifier, foo 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Errorf(tt.message, tt.args...).Error()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIs(t *testing.T) {
	t.Parallel()

	testErr := customError{message: "some error"}

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "nil is nil",
			err:    nil,
			target: nil,
			want:   true,
		},
		{
			name:   "nil error",
			err:    nil,
			target: testErr,
			want:   false,
		},
		{
			name:   "same error",
			err:    testErr,
			target: testErr,
			want:   true,
		},
		{
			name:   "different error",
			err:    testErr,
			target: errors.New("another error"),
			want:   false,
		},
		{
			name:   "with stack",
			err:    errors.WithStack(testErr),
			target: testErr,
			want:   true,
		},
		{
			name:   "with wrap",
			err:    errors.Wrap(testErr, "wrapped error"),
			target: testErr,
			want:   true,
		},
		{
			name:   "standard error compatibility",
			err:    fmt.Errorf("wrap with:%w", testErr),
			target: testErr,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Is(tt.err, tt.target)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	testErr := customError{message: "some error"}

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "nil error",
			err:  nil,
			want: nil,
		},
		{
			name: "non-wrapping error",
			err:  testErr,
			want: nil,
		},
		{
			name: "non-wrapping stackprune error",
			err:  errors.New("stackprune error"),
			want: nil,
		},
		{
			name: "with stack",
			err:  errors.WithStack(testErr),
			want: testErr,
		},
		{
			name: "with wrap",
			err:  errors.Wrap(testErr, "wrapped error"),
			want: testErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Unwrap(tt.err)

			assert.Equal(t, tt.want, got)
		})
	}
}
