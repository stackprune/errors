package errors_test

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/stackprune/errors/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestLogValue_ErrorTree(t *testing.T) {
	const wrappedJoinMessage = "context: first\nsecond"

	originalOptions := errors.GetLogOptions()

	t.Cleanup(func() { errors.SetLogOptions(originalOptions) })
	errors.SetLogOptions(errors.LogOptions{})

	first := errors.New("first")
	second := errors.New("second")
	third := errors.New("third")
	empty := errors.NewWithCallers("empty", nil)
	firstNode := logCause("first", testLogKindError)
	secondNode := logCause("second", testLogKindError)
	thirdNode := logCause("third", testLogKindError)
	joinedNode := logCause("first\nsecond", testLogKindJoinError, firstNode, secondNode)

	tests := []struct {
		name     string
		err      error
		message  string
		kind     string
		stack    error
		children []any
	}{
		{
			name: "stacked leaf", err: first,
			message: "first", kind: testLogKindError, stack: first,
		},
		{
			name: "repeated wrapping", err: errors.Wrap(errors.Wrap(first, "inner"), "outer"),
			message: "outer: inner: first", kind: testLogKindError, stack: first,
		},
		{
			name: "joined errors", err: errors.Join(first, second),
			message: "first\nsecond", kind: testLogKindJoinError, stack: first,
			children: []any{firstNode, secondNode},
		},
		{
			name: "wrapped children", err: errors.Join(errors.Wrap(first, "context"), second),
			message: wrappedJoinMessage, kind: testLogKindJoinError, stack: first,
			children: []any{logCause("context: first", testLogKindError), secondNode},
		},
		{
			name: "wrapped join", err: errors.Wrap(errors.Join(first, second), "context"),
			message: wrappedJoinMessage, kind: testLogKindJoinError, stack: first,
			children: []any{firstNode, secondNode},
		},
		{
			name: "nested join", err: errors.Join(errors.Join(first, second), third),
			message: "first\nsecond\nthird", kind: testLogKindJoinError, stack: first,
			children: []any{joinedNode, thirdNode},
		},
		{
			name:    "nested wrapped join",
			err:     errors.Join(first, errors.Wrap(errors.Join(second, third), "batch")),
			message: "first\nbatch: second\nthird",
			kind:    testLogKindJoinError,
			stack:   first,
			children: []any{
				firstNode,
				logCause("batch: second\nthird", testLogKindJoinError, secondNode, thirdNode),
			},
		},
		{
			name:     "standard wrapper around join",
			err:      errors.WithStack(fmt.Errorf("context: %w", errors.Join(first, second))),
			message:  wrappedJoinMessage,
			kind:     testLogKindJoinError,
			stack:    first,
			children: []any{firstNode, secondNode},
		},
		{
			name:    "standard join containing nested join",
			err:     errors.Join(stderrors.Join(errors.Join(first, second), third)),
			message: "first\nsecond\nthird",
			kind:    testLogKindJoinError,
			stack:   first,
			children: []any{
				logCause("first\nsecond\nthird", "*errors.joinError", joinedNode, thirdNode),
			},
		},
		{
			name:     "multiple percent-w operands",
			err:      errors.WithStack(fmt.Errorf("%w / %w", first, second)),
			message:  "first / second",
			kind:     "*fmt.wrapErrors",
			stack:    first,
			children: []any{firstNode, secondNode},
		},
		{
			name: "plain error before stack", err: errors.Join(io.EOF, second),
			message: "EOF\nsecond", kind: testLogKindJoinError, stack: second,
			children: []any{logCause("EOF", testLogKindStandardError), secondNode},
		},
		{
			name: "empty stack before usable stack", err: errors.Join(empty, second),
			message: "empty\nsecond", kind: testLogKindJoinError, stack: second,
			children: []any{logCause("empty", testLogKindError), secondNode},
		},
		{
			name:     "wrapped empty stack before usable stack",
			err:      errors.Wrap(errors.Join(empty, second), "context"),
			message:  "context: empty\nsecond",
			kind:     testLogKindJoinError,
			stack:    second,
			children: []any{logCause("empty", testLogKindError), secondNode},
		},
		{
			name: "no stack", err: errors.Join(io.EOF),
			message: "EOF", kind: testLogKindJoinError,
			children: []any{logCause("EOF", testLogKindStandardError)},
		},
		{
			name: "zero error", err: &errors.Error{},
			message: "", kind: testLogKindError,
		},
		{
			name: "zero join", err: &errors.JoinError{},
			message: "", kind: testLogKindJoinError,
		},
		{
			name:    "nil children",
			err:     errors.Join(&nilChildError{children: []error{nil, first, nil}}, second),
			message: "nullable\nsecond",
			kind:    testLogKindJoinError,
			stack:   first,
			children: []any{
				logCause("nullable", "*errors_test.nilChildError", firstNode),
				secondNode,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := []any{}
			if tt.stack != nil {
				stack = stackStrings(t, tt.stack)
			}

			children := tt.children
			if children == nil {
				children = []any{}
			}

			want, marshalErr := json.Marshal(map[string]any{
				testLogKeyMessage:  tt.message,
				testLogKeyKind:     tt.kind,
				testLogKeyStack:    stack,
				testLogKeyChildren: children,
			})
			require.NoError(t, marshalErr)

			var output bytes.Buffer

			slog.New(slog.NewJSONHandler(&output, nil)).Error("failed", slog.Any("error", tt.err))

			var record map[string]json.RawMessage

			require.NoError(t, json.Unmarshal(output.Bytes(), &record))
			assert.JSONEq(t, string(want), string(record["error"]))
		})
	}
}

//nolint:paralleltest
func TestLogValue_Handlers(t *testing.T) {
	originalOptions := errors.GetLogOptions()

	t.Cleanup(func() { errors.SetLogOptions(originalOptions) })

	first := errors.New("first")
	second := errors.New("second")
	err := errors.Wrap(errors.Join(first, second), "outer")

	var stackErr *errors.Error

	require.ErrorAs(t, first, &stackErr)

	for _, format := range []errors.StackFormat{
		errors.StackFormatStringArray,
		errors.StackFormatObjectArray,
		errors.StackFormat(-1),
	} {
		t.Run(fmt.Sprintf("format=%d", format), func(t *testing.T) {
			errors.SetLogOptions(errors.LogOptions{
				MessageKey:  testLogKeyCustomMessage,
				KindKey:     testLogKeyCustomKind,
				StackKey:    testLogKeyCustomStack,
				StackFormat: format,
			})

			stack := stackStrings(t, first)

			if format == errors.StackFormatObjectArray {
				frames := stackErr.Stacks()
				stack = make([]any, 0, len(frames))

				for _, frame := range frames {
					stack = append(stack, map[string]any{
						"function": frame.FuncName,
						"file":     frame.File,
						"line":     frame.LineNumber,
					})
				}
			}

			want, marshalErr := json.Marshal(map[string]any{
				testLogKeyCustomMessage: "outer: first\nsecond",
				testLogKeyCustomKind:    testLogKindJoinError,
				testLogKeyCustomStack:   stack,
				testLogKeyChildren: []any{
					map[string]any{
						testLogKeyCustomMessage: "first",
						testLogKeyCustomKind:    testLogKindError,
						testLogKeyChildren:      []any{},
					},
					map[string]any{
						testLogKeyCustomMessage: "second",
						testLogKeyCustomKind:    testLogKindError,
						testLogKeyChildren:      []any{},
					},
				},
			})
			require.NoError(t, marshalErr)

			var jsonOutput bytes.Buffer

			slog.New(slog.NewJSONHandler(&jsonOutput, nil)).Error("failed", slog.Any("error", err))

			var record map[string]json.RawMessage

			require.NoError(t, json.Unmarshal(jsonOutput.Bytes(), &record))
			assert.JSONEq(t, string(want), string(record["error"]))

			var textOutput bytes.Buffer

			slog.New(slog.NewTextHandler(&textOutput, nil)).Error("failed", slog.Any("error", err))
			text := textOutput.String()
			assert.Contains(
				t,
				text,
				"error."+testLogKeyCustomMessage+"="+strconv.Quote("outer: first\nsecond"),
			)
			assert.Contains(t, text, "error."+testLogKeyCustomKind+"=*errors.JoinError")
			assert.Equal(t, 1, strings.Count(text, "error."+testLogKeyCustomStack+"="))
			assert.Contains(t, text, "error.children=")
			assert.Contains(t, text, testLogKeyCustomMessage+":first")
			assert.Contains(t, text, testLogKeyCustomMessage+":second")
			assert.NotContains(t, text, testLogKeyCustomStack+":")
		})
	}
}

func TestJoinError_LogValue_PreservesErrors(t *testing.T) {
	t.Parallel()

	first := errors.New("first")
	second := errors.New("second")
	joined := errors.Join(first, second)
	firstStack := stackStrings(t, first)
	secondStack := stackStrings(t, second)
	verbose := fmt.Sprintf("%+v", joined)

	var joinErr *errors.JoinError

	require.ErrorAs(t, joined, &joinErr)
	_ = joinErr.LogValue()

	assert.Equal(t, []error{first, second}, joinErr.Unwrap())
	assert.Equal(t, firstStack, stackStrings(t, first))
	assert.Equal(t, secondStack, stackStrings(t, second))
	assert.Equal(t, verbose, fmt.Sprintf("%+v", joined))
	require.ErrorIs(t, joined, first)
	require.ErrorIs(t, joined, second)
}

func logCause(message, kind string, children ...any) map[string]any {
	if children == nil {
		children = []any{}
	}

	return map[string]any{
		testLogKeyMessage:  message,
		testLogKeyKind:     kind,
		testLogKeyChildren: children,
	}
}

type nilChildError struct {
	children []error
}

func (e *nilChildError) Error() string {
	return "nullable"
}

func (e *nilChildError) Unwrap() []error {
	return e.children
}
