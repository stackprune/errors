package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/stackprune/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoin(t *testing.T) {
	t.Parallel()

	err1 := errors.New("first error")
	err2 := errors.New("second error")
	standardErr := stderrors.New("standard error")

	tests := []struct {
		name string
		errs []error
		want []error
	}{
		{
			name: "errors.Join()",
			errs: []error{},
			want: nil,
		},
		{
			name: "errors.Join(nil, nil)",
			errs: []error{nil, nil},
			want: nil,
		},
		{
			name: "single error in list",
			errs: []error{err1},
			want: []error{err1},
		},
		{
			name: "multiple errors in list",
			errs: []error{err1, err2},
			want: []error{err1, err2},
		},
		{
			name: "list with nil errors",
			errs: []error{err1, nil, err2},
			want: []error{err1, err2},
		},
		{
			name: "list with stackprune and standard errors",
			errs: []error{err1, standardErr},
			want: []error{err1, standardErr},
		},
		{
			name: "list with standard and stackprune errors",
			errs: []error{standardErr, err1},
			want: []error{standardErr, err1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := errors.Join(tt.errs...)
			if tt.want == nil {
				assert.NoError(t, got)

				return
			}

			require.Error(t, got)

			joinError := &errors.JoinError{}
			require.ErrorAs(t, got, &joinError)
			assert.Equal(t, tt.want, joinError.Unwrap())
		})
	}
}
