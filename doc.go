// Package errors provides a modern, lightweight alternative to pkg/errors,
// capturing a single stack trace at error creation time and preserving it
// through all wraps. It avoids redundant stack traces while maintaining
// compatibility with Go 1.13+ errors.Is / errors.As / errors.Unwrap.
//
// The API includes `New`, `Wrap`, `Errorf`, `WithStack`, and `Join`, offering
// familiar ergonomics without the noise of `WithMessage` or similar.
package errors
