// Package errors provides a modern, lightweight alternative to pkg/errors,
// capturing a single stack trace at error creation time and preserving it
// through all wraps. It avoids redundant stack traces while maintaining
// compatibility with Go 1.13+ errors.Is / errors.As / errors.Unwrap.
//
// Import this package as "github.com/stackprune/errors/v2".
//
// The API includes `New`, `Wrap`, `Errorf`, `WithStack`, and `Join`, offering
// familiar ergonomics without the noise of `WithMessage` or similar.
//
// Both *Error and *JoinError implement slog.LogValuer with a common message,
// kind, stack, and children structure. Only the first available stack is logged,
// at the top level. Joined errors retain their grouping, while wrapping context
// is combined into each message without adding a level to the logged tree.
package errors
