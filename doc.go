// Package errors provides a modern, lightweight alternative to pkg/errors,
// capturing a single stack trace at error creation time and preserving it
// through all wraps. It avoids redundant stack traces while maintaining
// compatibility with Go 1.13+ errors.Is / errors.As / errors.Unwrap.
//
// Version 2 is available at https://pkg.go.dev/github.com/stackprune/errors/v2.
// This package retains v1 behavior. For migration guidance, see
// https://github.com/stackprune/errors/tree/v2.0.0#upgrading-from-v1.
//
// The API includes `New`, `Wrap`, `Errorf`, `WithStack`, and `Join`, offering
// familiar ergonomics without the noise of `WithMessage` or similar.
//
// The *Error type also implements slog.LogValuer for structured logging.
package errors
