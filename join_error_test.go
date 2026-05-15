package errors_test

import (
	stderrors "errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/stackprune/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		errs []error
		want string
	}{
		{
			name: "empty error list",
			errs: []error{},
			want: "",
		},
		{
			name: "single error",
			errs: []error{stderrors.New("single error")},
			want: "single error",
		},
		{
			name: "multiple errors",
			errs: []error{
				stderrors.New("first error"),
				stderrors.New("second error"),
			},
			want: "first error\nsecond error",
		},
		{
			name: "errors with nil values",
			errs: []error{
				stderrors.New("first error"),
				nil,
				stderrors.New("third error"),
			},
			want: "first error\nthird error",
		},
		{
			name: "mixed error types",
			errs: []error{
				errors.New("stackprune error"),
				stderrors.New("standard error"),
			},
			want: "stackprune error\nstandard error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Join(tt.errs...)
			if tt.want == "" {
				assert.NoError(t, got)

				return
			}

			require.Error(t, got)
			assert.Equal(t, tt.want, got.Error())
		})
	}
}

func TestJoinError_Format(t *testing.T) {
	t.Parallel()

	err1 := errors.New("first stackprune error")
	err2 := errors.New("second stackprune error")
	standardErr1 := stderrors.New("first error")
	standardErr2 := stderrors.New("second error")

	tests := []struct {
		name       string
		joinedErr  error
		format     string
		want       string
		wantRegexp *regexp.Regexp
	}{
		{
			name:      "simple string format",
			joinedErr: errors.Join(err1, standardErr2),
			format:    "%s",
			want:      "first stackprune error\nsecond error",
		},
		{
			name:      "simple string format with multiple errors",
			joinedErr: errors.Join(standardErr1, standardErr2, err1),
			format:    "%s",
			want:      "first error\nsecond error\nfirst stackprune error",
		},
		{
			name:      "quoted format with escaped newlines",
			joinedErr: errors.Join(err1, standardErr2),
			format:    "%q",
			want:      `"first stackprune error\nsecond error"`,
		},
		{
			name:      "simple verbose format same as string",
			joinedErr: errors.Join(err1, standardErr2),
			format:    "%v",
			want:      "first stackprune error\nsecond error",
		},
		{
			name:      "detailed verbose format with stack trace",
			joinedErr: errors.Join(standardErr1, err2),
			format:    testFormatDetailedVerbose,
			wantRegexp: regexp.MustCompile(
				`(?s)^first error\n` +
					`second stackprune error\n` +
					`github\.com/stackprune/errors_test\.TestJoinError_Format\n` +
					`\t[\S]+/join_error_test\.go:\d+\n`,
			),
		},
		{
			name:      "verbose format with mixed error types and stack traces",
			joinedErr: errors.Join(standardErr1, err1, err2),
			format:    testFormatDetailedVerbose,
			wantRegexp: regexp.MustCompile(
				`(?s)^first error\n` +
					`first stackprune error\n` +
					`github\.com/stackprune/errors_test\.TestJoinError_Format\n` +
					`\t[\S]+/join_error_test\.go:\d+\n` +
					`.+\n` +
					`second stackprune error`,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := fmt.Sprintf(tt.format, tt.joinedErr)
			if tt.wantRegexp == nil {
				assert.Equal(t, tt.want, got)
			} else {
				assert.Regexp(t, tt.wantRegexp, got)
			}
		})
	}
}

func TestJoinError_Unwrap(t *testing.T) {
	t.Parallel()

	standardErr1 := stderrors.New("first error")
	standardErr2 := stderrors.New("second error")
	err := errors.New("third error")
	joinedStandardErr := errors.Join(standardErr1, standardErr2)

	tests := []struct {
		name      string
		joinedErr error
		want      []error
	}{
		{
			name:      "unwrap multiple errors to slice",
			joinedErr: errors.Join(standardErr1, standardErr2, err),
			want:      []error{standardErr1, standardErr2, err},
		},
		{
			name:      "unwrap single error to slice",
			joinedErr: errors.Join(standardErr1),
			want:      []error{standardErr1},
		},
		{
			name:      "unwrap multiple errors with joined standard error",
			joinedErr: errors.Join(joinedStandardErr, err),
			want:      []error{joinedStandardErr, err},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := func() *errors.JoinError {
				var target *errors.JoinError

				_ = errors.As(tt.joinedErr, &target)

				return target
			}().Unwrap()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJoinError_Is(t *testing.T) {
	t.Parallel()

	standardErr1 := stderrors.New("first error")
	standardErr2 := stderrors.New("second error")
	unrelatedErr := stderrors.New("third error")
	err := errors.New("stackprune error")

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "should match first error in joined errors",
			err:    errors.Join(standardErr1, standardErr2),
			target: standardErr1,
			want:   true,
		},
		{
			name:   "should match second error in joined errors",
			err:    errors.Join(standardErr1, standardErr2),
			target: standardErr2,
			want:   true,
		},
		{
			name:   "should not match unrelated error",
			err:    errors.Join(standardErr1, standardErr2),
			target: unrelatedErr,
			want:   false,
		},
		{
			name:   "should match stackprune error in mixed joined errors",
			err:    errors.Join(standardErr1, err, standardErr2),
			target: err,
			want:   true,
		},
		{
			name:   "should match error in single-error join",
			err:    errors.Join(standardErr1),
			target: standardErr1,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.want {
				require.ErrorIs(t, tt.err, tt.target)
			} else {
				assert.NotErrorIs(t, tt.err, tt.target)
			}
		})
	}
}

func TestJoinError_As(t *testing.T) {
	t.Parallel()

	customErr := customError{message: "custom error"}
	standardErr1 := stderrors.New("standard error")
	err := errors.New("stackprune error")

	tests := []struct {
		name      string
		joinedErr error
		testFunc  func(*testing.T, error)
	}{
		{
			name:      "should extract custom error type from joined errors",
			joinedErr: errors.Join(customErr, standardErr1),
			testFunc: func(t *testing.T, joinedErr error) {
				t.Helper()

				var target customError
				require.ErrorAs(t, joinedErr, &target)
				assert.Equal(t, "custom error", target.message)
			},
		},
		{
			name:      "should extract JoinError type itself",
			joinedErr: errors.Join(customErr, standardErr1),
			testFunc: func(t *testing.T, joinedErr error) {
				t.Helper()

				var joinErr *errors.JoinError
				require.ErrorAs(t, joinedErr, &joinErr)
				assert.NotNil(t, joinErr)
			},
		},
		{
			name:      "should extract stackprune Error type",
			joinedErr: errors.Join(err, standardErr1),
			testFunc: func(t *testing.T, joinedErr error) {
				t.Helper()

				var stackpruneErr *errors.Error
				require.ErrorAs(t, joinedErr, &stackpruneErr)
				assert.NotNil(t, stackpruneErr)
				assert.Equal(t, "stackprune error", stackpruneErr.Error())
			},
		},
		{
			name:      "should not find non-existent custom error type",
			joinedErr: errors.Join(standardErr1),
			testFunc: func(t *testing.T, joinedErr error) {
				t.Helper()

				var target customError
				assert.NotErrorAs(t, joinedErr, &target)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.testFunc(t, tt.joinedErr)
		})
	}
}
