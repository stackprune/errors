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

const (
	defaultMessageKey = "message"
	defaultKindKey    = "kind"
	defaultStackKey   = "stack"
	defaultErrorsKey  = "errors"
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
			MessageKey:  defaultMessageKey,
			KindKey:     defaultKindKey,
			StackKey:    defaultStackKey,
			StackFormat: StackFormatStringArray,
		}
	}

	return opt
}

// SetLogOptions updates the global LogOptions used for logging *Error values with slog.
// This affects all future calls to *Error.LogValue.
// This function is safe for concurrent use.
func SetLogOptions(options LogOptions) {
	if options.MessageKey == "" {
		options.MessageKey = defaultMessageKey
	}

	if options.KindKey == "" {
		options.KindKey = defaultKindKey
	}

	if options.StackKey == "" {
		options.StackKey = defaultStackKey
	}

	logOptionsValue.Store(options)
}

// LogValue implements slog.LogValuer for *Error.
// It returns a structured slog.Value containing the error kind, message, and stack trace.
// Stack trace formatting is determined by the global LogOptions.
func (e *Error) LogValue() slog.Value {
	logOptions := GetLogOptions()

	return slog.GroupValue(
		slog.String(logOptions.MessageKey, e.Error()),
		slog.String(logOptions.KindKey, rootErrorKind(e)),
		slog.Any(logOptions.StackKey, formatStackItems(e.Stacks(), logOptions.StackFormat)),
	)
}

// LogValue implements slog.LogValuer for *JoinError.
// It returns a structured value containing the joined message and each child error.
func (j *JoinError) LogValue() slog.Value {
	logOptions := GetLogOptions()
	items := make([]any, 0, len(j.errs))

	for _, err := range j.errs {
		if err != nil {
			items = append(items, logErrorMap(err, logOptions))
		}
	}

	return slog.GroupValue(
		slog.String(logOptions.MessageKey, j.Error()),
		slog.String(logOptions.KindKey, reflect.TypeFor[*JoinError]().String()),
		slog.Any(defaultErrorsKey, items),
	)
}

func logErrorMap(err error, logOptions LogOptions) map[string]any {
	result := map[string]any{
		logOptions.MessageKey: err.Error(),
		logOptions.KindKey:    reflect.TypeOf(err).String(),
	}

	var joinErr *JoinError
	if As(err, &joinErr) {
		items := make([]any, 0, len(joinErr.errs))

		for _, child := range joinErr.errs {
			if child != nil {
				items = append(items, logErrorMap(child, logOptions))
			}
		}

		result[defaultErrorsKey] = items

		return result
	}

	var errorWithStack *Error
	if As(err, &errorWithStack) {
		result[logOptions.KindKey] = rootErrorKind(errorWithStack)
		result[logOptions.StackKey] = formatStackItems(
			errorWithStack.Stacks(),
			logOptions.StackFormat,
		)
	}

	return result
}

func rootErrorKind(e *Error) string {
	rootErr := error(e)

	for {
		unwrapped := Unwrap(rootErr)
		if unwrapped == nil {
			break
		}

		rootErr = unwrapped
	}

	return reflect.TypeOf(rootErr).String()
}

func formatStackItems(stackFrames []Stack, stackFormat StackFormat) []any {
	stackItems := make([]any, 0, len(stackFrames))

	switch stackFormat {
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

	return stackItems
}

var _ slog.LogValuer = (*Error)(nil)
var _ slog.LogValuer = (*JoinError)(nil)
