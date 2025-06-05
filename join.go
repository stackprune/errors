// join.go defines the Join function to combine multiple error values.
// It wraps Go 1.20+ errors.Join where available, and provides a fallback
// that preserves stack-aware behavior across combined errors.

package errors

// Join combines multiple errors into a single error value, ignoring nils.
func Join(errs ...error) error {
	nonNilErrs := make([]error, 0, len(errs))

	for _, e := range errs {
		if e != nil {
			nonNilErrs = append(nonNilErrs, e)
		}
	}

	if len(nonNilErrs) == 0 {
		return nil
	}

	return &JoinError{errs: nonNilErrs}
}
