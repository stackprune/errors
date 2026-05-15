package errors_test

import (
	"io"
	"log/slog"
	"strconv"
	"testing"

	"github.com/stackprune/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestSetLogOptions(t *testing.T) {
	// Save original options to restore later
	defaultOptions := errors.GetLogOptions()

	// Restore original options after test
	t.Cleanup(func() {
		errors.SetLogOptions(defaultOptions)
	})

	customOptions := errors.LogOptions{
		KindKey:     "error_type",
		MessageKey:  "error_message",
		StackKey:    "stack_trace",
		StackFormat: errors.StackFormatObjectArray,
	}

	errors.SetLogOptions(customOptions)

	currentOptions := errors.GetLogOptions()

	assert.Equal(t, customOptions, currentOptions)
}

//nolint:paralleltest
func TestSetLogOptions_DefaultHandling(t *testing.T) {
	defaultOptions := errors.GetLogOptions()

	t.Cleanup(func() {
		errors.SetLogOptions(defaultOptions)
	})

	tests := []struct {
		name  string
		input errors.LogOptions
		want  errors.LogOptions
	}{
		{
			name: "empty MessageKey uses default",
			input: errors.LogOptions{
				MessageKey:  "", // empty
				KindKey:     "custom_kind",
				StackKey:    "custom_stack",
				StackFormat: errors.StackFormatObjectArray,
			},
			want: errors.LogOptions{
				MessageKey:  "message", // default
				KindKey:     "custom_kind",
				StackKey:    "custom_stack",
				StackFormat: errors.StackFormatObjectArray,
			},
		},
		{
			name: "empty KindKey uses default",
			input: errors.LogOptions{
				MessageKey:  "custom_message",
				KindKey:     "", // empty
				StackKey:    "custom_stack",
				StackFormat: errors.StackFormatObjectArray,
			},
			want: errors.LogOptions{
				MessageKey:  "custom_message",
				KindKey:     "kind", // default
				StackKey:    "custom_stack",
				StackFormat: errors.StackFormatObjectArray,
			},
		},
		{
			name: "empty StackKey uses default",
			input: errors.LogOptions{
				MessageKey:  "custom_message",
				KindKey:     "custom_kind",
				StackKey:    "", // empty
				StackFormat: errors.StackFormatObjectArray,
			},
			want: errors.LogOptions{
				MessageKey:  "custom_message",
				KindKey:     "custom_kind",
				StackKey:    "stack", // default
				StackFormat: errors.StackFormatObjectArray,
			},
		},
		{
			name: "all empty keys use defaults",
			input: errors.LogOptions{
				MessageKey:  "", // empty
				KindKey:     "", // empty
				StackKey:    "", // empty
				StackFormat: errors.StackFormatObjectArray,
			},
			want: errors.LogOptions{
				MessageKey:  "message", // default
				KindKey:     "kind",    // default
				StackKey:    "stack",   // default
				StackFormat: errors.StackFormatObjectArray,
			},
		},
		{
			name: "non-empty keys preserved",
			input: errors.LogOptions{
				MessageKey:  "error_msg",
				KindKey:     "error_kind",
				StackKey:    "error_stack",
				StackFormat: errors.StackFormatStringArray,
			},
			want: errors.LogOptions{
				MessageKey:  "error_msg",
				KindKey:     "error_kind",
				StackKey:    "error_stack",
				StackFormat: errors.StackFormatStringArray,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors.SetLogOptions(tt.input)
			got := errors.GetLogOptions()

			assert.Equal(t, tt.want, got)
		})
	}
}

//nolint:paralleltest
func TestError_LogValue(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		stackFormat errors.StackFormat
		check       func(t *testing.T, message string, kind string, stackFrames []any)
	}{
		{
			name:        "stack format: string array",
			err:         errors.New("test error"),
			stackFormat: errors.StackFormatStringArray,
			check: func(t *testing.T, message string, kind string, stackFrames []any) {
				t.Helper()

				assert.Equal(t, "test error", message)
				assert.Equal(t, "*errors.Error", kind)

				for _, frame := range stackFrames {
					stackString, ok := frame.(string)
					require.True(t, ok, "Stack item should be a string")
					assert.Regexp(
						t,
						`^[\S]+ at [\S]+\.go:\d+$`,
						stackString,
						"Stack item should match expected format 'function at file:line'",
					)
				}
			},
		},
		{
			name:        "stack format: object array",
			err:         errors.New("test error"),
			stackFormat: errors.StackFormatObjectArray,
			check: func(t *testing.T, message string, kind string, stackFrames []any) {
				t.Helper()

				assert.Equal(t, "test error", message)
				assert.Equal(t, "*errors.Error", kind)

				for _, item := range stackFrames {
					stack, ok := item.(map[string]any)

					require.True(t, ok, "Stack item should be a map[string]any")
					assert.IsType(t, "", stack["function"],
						"Stack item 'function' should be string",
					)
					assert.IsType(t, "", stack["file"],
						"Stack item 'file' should be string",
					)
					assert.IsType(t, 0, stack["line"],
						"Stack item 'line' should be int",
					)
				}
			},
		},
		{
			name:        "wrapped error",
			err:         errors.Wrap(io.EOF, "wrapped error"),
			stackFormat: errors.StackFormatStringArray,
			check: func(t *testing.T, message string, kind string, stackFrames []any) {
				t.Helper()

				assert.Equal(t, "wrapped error: EOF", message)
				assert.Equal(t, "*errors.errorString", kind)

				for _, frame := range stackFrames {
					stackString, ok := frame.(string)
					require.True(t, ok, "Stack item should be a string")
					assert.Regexp(
						t,
						`^[\S]+ at [\S]+\.go:\d+$`,
						stackString,
						"Stack item should match expected format 'function at file:line'",
					)
				}
			},
		},
		{
			name:        "empty error",
			err:         errors.New(""),
			stackFormat: errors.StackFormatStringArray,
			check: func(t *testing.T, message string, kind string, stackFrames []any) {
				t.Helper()

				assert.Empty(t, message)
				assert.Equal(t, "*errors.Error", kind)
				assert.NotEmpty(t, stackFrames)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set stack format
			defaultOptions := errors.GetLogOptions()

			t.Cleanup(func() {
				errors.SetLogOptions(defaultOptions)
			})
			errors.SetLogOptions(errors.LogOptions{
				MessageKey:  "message",
				KindKey:     "kind",
				StackKey:    "stack",
				StackFormat: tt.stackFormat,
			})

			baseErr := tt.err

			var err *errors.Error

			require.ErrorAs(t, baseErr, &err)

			group := err.LogValue().Group()
			require.Len(t, group, 3)

			messageAttr := group[0]
			kindAttr := group[1]
			stackAttr := group[2]

			// Check structure
			require.Equal(t, "message", messageAttr.Key)
			require.Equal(t, "kind", kindAttr.Key)
			require.Equal(t, "stack", stackAttr.Key)

			messageValue := messageAttr.Value.String()
			kindValue := kindAttr.Value.String()
			stackValue := stackAttr.Value.Any()
			stackSlice, ok := stackValue.([]any)
			require.True(t, ok, "stack should be a []any")

			if tt.check != nil {
				tt.check(t, messageValue, kindValue, stackSlice)
			}
		})
	}
}

//nolint:paralleltest
func TestJoinError_LogValue(t *testing.T) {
	defaultOptions := errors.GetLogOptions()

	t.Cleanup(func() {
		errors.SetLogOptions(defaultOptions)
	})
	errors.SetLogOptions(errors.LogOptions{
		MessageKey:  "message",
		KindKey:     "kind",
		StackKey:    "stack",
		StackFormat: errors.StackFormatStringArray,
	})

	err1 := errors.New("first error")
	err2 := errors.New("second error")
	joinedErr := errors.Join(err1, err2)

	var joinErr *errors.JoinError
	require.ErrorAs(t, joinedErr, &joinErr)

	group := joinErr.LogValue().Group()

	assert.Equal(t, []slog.Attr{
		slog.String("message", "first error\nsecond error"),
		slog.String("kind", "*errors.JoinError"),
		slog.Any("errors", []any{
			map[string]any{
				"message": "first error",
				"kind":    "*errors.Error",
				"stack":   stackStrings(t, err1),
			},
			map[string]any{
				"message": "second error",
				"kind":    "*errors.Error",
				"stack":   stackStrings(t, err2),
			},
		}),
	}, group)
}

func stackStrings(t *testing.T, err error) []any {
	t.Helper()

	var stackErr *errors.Error
	require.ErrorAs(t, err, &stackErr)

	stackItems := make([]any, 0, len(stackErr.Stacks()))

	for _, frame := range stackErr.Stacks() {
		stackItems = append(
			stackItems,
			frame.FuncName+" at "+frame.File+":"+strconv.Itoa(frame.LineNumber),
		)
	}

	return stackItems
}
