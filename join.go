// join.go defines the Join function to combine multiple error values.
// It returns a JoinError that preserves its children for unwrapping,
// formatting, and structured logging.

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
