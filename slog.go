// slog.go defines slog.LogValuer integration for *Error with configurable stack trace formatting.
// It enables structured logging using slog, allowing customization of field names and output format.

package errors

import (
	"log/slog"
	"reflect"
	"strconv"
	"sync/atomic"
)

// StackFormat specifies the output format of stack traces in logs.
type StackFormat int

const (
	// StackFormatStringArray renders the stack trace as an array of strings.
	StackFormatStringArray StackFormat = iota

	// StackFormatObjectArray renders the stack trace as an array of structured objects.
	StackFormatObjectArray
)

// LogOptions defines how *Error values are serialized for slog structured logging.
// It allows customization of key names and stack trace formatting.
type LogOptions struct {
	MessageKey  string
	KindKey     string
	StackKey    string
	StackFormat StackFormat
}

//nolint:gochecknoglobals
var logOptionsValue atomic.Value

// GetLogOptions returns the current global LogOptions.
// If not explicitly set, default values are used.
// This function is safe for concurrent use.
func GetLogOptions() LogOptions {
	v := logOptionsValue.Load()

	opt, ok := v.(LogOptions)
	if !ok {
		return LogOptions{
			MessageKey:  "message",
			KindKey:     "kind",
			StackKey:    "stack",
			StackFormat: StackFormatStringArray,
		}
	}

	return opt
}

// SetLogOptions updates the global LogOptions used for logging *Error values with slog.
// This affects all future calls to *Error.LogValue.
// This function is safe for concurrent use.
func SetLogOptions(o LogOptions) {
	logOptionsValue.Store(o)
}

// LogValue implements slog.LogValuer for *Error.
// It returns a structured slog.Value containing the error kind, message, and stack trace.
// Stack trace formatting is determined by the global LogOptions.
func (e *Error) LogValue() slog.Value {
	logOptions := GetLogOptions()
	stackFrames := e.Stacks()
	stackItems := make([]any, 0, len(stackFrames))

	switch logOptions.StackFormat {
	case StackFormatObjectArray:
		for _, frame := range stackFrames {
			stackItems = append(stackItems, map[string]any{
				"function": frame.FuncName,
				"file":     frame.File,
				"line":     frame.LineNumber,
			})
		}
	case StackFormatStringArray:
		fallthrough // fall back to default for legacy and unknown formats
	default:
		for _, frame := range stackFrames {
			stackItems = append(
				stackItems,
				frame.FuncName+" at "+frame.File+":"+strconv.Itoa(frame.LineNumber),
			)
		}
	}

	rootErr := error(e)

	for {
		unwrapped := Unwrap(rootErr)
		if unwrapped == nil {
			break
		}

		rootErr = unwrapped
	}

	return slog.GroupValue(
		slog.String(logOptions.MessageKey, e.Error()),
		slog.String(logOptions.KindKey, reflect.TypeOf(rootErr).String()),
		slog.Any(logOptions.StackKey, stackItems),
	)
}

var _ slog.LogValuer = (*Error)(nil)
