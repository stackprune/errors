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
	defaultMessageKey  = "message"
	defaultKindKey     = "kind"
	defaultStackKey    = "stack"
	defaultChildrenKey = "children"
)

// LogOptions defines how *Error and *JoinError values are serialized for slog structured logging.
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

// SetLogOptions updates the global LogOptions used for logging *Error and *JoinError values with slog.
// This affects all future calls to their LogValue methods.
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
// It returns message, kind, stack, and children fields using the global LogOptions.
// Wrap adds context to the message without adding a level to the error tree.
func (e *Error) LogValue() slog.Value {
	return logErrorValue(e)
}

// LogValue implements slog.LogValuer for *JoinError.
// It preserves the joined error tree and emits only the first available stack
// at the top level. Child errors contain message, kind, and children fields.
func (j *JoinError) LogValue() slog.Value {
	return logErrorValue(j)
}

func logErrorValue(err error) slog.Value {
	logOptions := GetLogOptions()

	return slog.GroupValue(
		slog.String(logOptions.MessageKey, err.Error()),
		slog.String(logOptions.KindKey, rootErrorKind(err)),
		slog.Any(logOptions.StackKey, formatStackItems(firstStack(err), logOptions.StackFormat)),
		slog.Any(defaultChildrenKey, logErrorChildren(err, logOptions)),
	)
}

func logErrorMap(err error, logOptions LogOptions) map[string]any {
	return map[string]any{
		logOptions.MessageKey: err.Error(),
		logOptions.KindKey:    rootErrorKind(err),
		defaultChildrenKey:    logErrorChildren(err, logOptions),
	}
}

func logErrorChildren(err error, logOptions LogOptions) []any {
	for {
		if joined, ok := err.(interface{ Unwrap() []error }); ok {
			children := joined.Unwrap()
			items := make([]any, 0, len(children))

			for _, child := range children {
				if child != nil {
					items = append(items, logErrorMap(child, logOptions))
				}
			}

			return items
		}

		err = Unwrap(err)
		if err == nil {
			return []any{}
		}
	}
}

// firstStack traverses errors in depth-first order without decoding later stacks.
//
//nolint:errorlint // Inspect the current node; errors.As would also search descendants.
func firstStack(err error) []Stack {
	if stackErr, ok := err.(*Error); ok {
		if stack := stackErr.Stacks(); len(stack) > 0 {
			return stack
		}
	}

	switch wrapped := err.(type) {
	case interface{ Unwrap() []error }:
		for _, child := range wrapped.Unwrap() {
			if stack := firstStack(child); len(stack) > 0 {
				return stack
			}
		}
	case interface{ Unwrap() error }:
		return firstStack(wrapped.Unwrap())
	}

	return nil
}

func rootErrorKind(err error) string {
	rootErr := err

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
