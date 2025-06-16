package errors_test

import (
	stderrors "errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/stackprune/errors"
)

// ExampleNew demonstrates basic error creation with stack trace.
func ExampleNew() {
	err := errors.New("something went wrong")
	fmt.Println(err.Error())
	// Output: something went wrong
}

// ExampleErrorf demonstrates formatted error creation.
func ExampleErrorf() {
	userID := 42
	err := errors.Errorf("user %d not found", userID)
	fmt.Println(err.Error())
	// Output: user 42 not found
}

// ExampleWrap demonstrates wrapping an existing error.
func ExampleWrap() {
	originalErr := stderrors.New("connection failed")
	wrappedErr := errors.Wrap(originalErr, "database access failed")
	fmt.Println(wrappedErr.Error())
	// Output: database access failed: connection failed
}

// ExampleWrapf demonstrates formatted wrapping of an error.
func ExampleWrapf() {
	originalErr := stderrors.New("timeout")
	tableName := "users"
	wrappedErr := errors.Wrapf(originalErr, "failed to query table %s", tableName)
	fmt.Println(wrappedErr.Error())
	// Output: failed to query table users: timeout
}

// ExampleWithStack demonstrates adding stack trace to standard errors.
func ExampleWithStack() {
	standardErr := stderrors.New("standard library error")
	stackErr := errors.WithStack(standardErr)
	fmt.Println(stackErr.Error())
	// Output: standard library error
}

// ExampleJoin demonstrates joining multiple errors.
func ExampleJoin() {
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	err3 := stderrors.New("third error")

	joinedErr := errors.Join(err1, err2, err3)
	fmt.Println(joinedErr.Error())
	// Output: first error
	// second error
	// third error
}

// ExampleJoin_withNil demonstrates that nil errors are ignored in Join.
func ExampleJoin_withNil() {
	err1 := errors.New("first error")
	err2 := errors.New("second error")

	joinedErr := errors.Join(err1, nil, err2, nil)
	fmt.Println(joinedErr.Error())
	// Output: first error
	// second error
}

// ExampleJoin_single demonstrates that joining a single error returns the original error.
func ExampleJoin_single() {
	originalErr := errors.New("single error")
	joinedErr := errors.Join(originalErr)

	fmt.Printf("Original: %s\n", originalErr)
	fmt.Printf("Joined: %s\n", joinedErr)

	// Output: Original: single error
	// Joined: single error
}

// ExampleUnwrap demonstrates unwrapping errors.
func ExampleUnwrap() {
	baseErr := errors.New("base error")
	wrappedErr := errors.Wrap(baseErr, "wrapped error")

	unwrapped := errors.Unwrap(wrappedErr)
	fmt.Println(unwrapped.Error())

	// Output: base error
}

// ExampleError_Format demonstrates different formatting options.
func ExampleError_Format() {
	err := errors.New("example error")

	// Simple string format
	fmt.Printf("%%s: %s\n", err)

	// Quoted format
	fmt.Printf("%%q: %q\n", err)

	// Simple verbose format
	fmt.Printf("%%v: %v\n", err)

	// Output: %s: example error
	// %q: "example error"
	// %v: example error
}

func handlerCreateUser() error {
	if err := usecaseCreateUser(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func usecaseCreateUser() error {
	if err := repositoryInsertUser(); err != nil {
		return errors.Wrap(err, "user creation failed")
	}

	return nil
}

func repositoryInsertUser() error {
	return errors.New("failed to insert user into database")
}

// ExampleError_Format_stackTrace demonstrates stack trace formatting with %+v.
// across handler → usecase → repository layers using Wrap and WithStack.
//
//nolint:testableexamples
func ExampleError_Format_stackTrace() {
	err := handlerCreateUser()

	fmt.Printf("Error: %+v\n", err)

	// Example output:
	// Error: user creation failed: failed to insert user into database
	// github.com/stackprune/errors_test.repositoryInsertUser
	// 	/app/example_test.go:133
	// github.com/stackprune/errors_test.usecaseCreateUser
	// 	/app/example_test.go:125
	// github.com/stackprune/errors_test.handlerCreateUser
	// 	/app/example_test.go:117
	// github.com/stackprune/errors_test.ExampleError_Format_stackTrace
	// 	/app/example_test.go:139
}

// ExampleWrap_networkError demonstrates wrapping real-world errors like network timeouts.
func ExampleWrap_networkError() {
	// Simulate a network error (like io.ErrUnexpectedEOF)
	networkErr := io.ErrUnexpectedEOF

	// Wrap with context
	connectionErr := errors.Wrap(networkErr, "network connection lost")
	serviceErr := errors.Wrapf(connectionErr, "failed to fetch data from %s", "api.example.com")

	fmt.Println(serviceErr.Error())

	// Output: failed to fetch data from api.example.com: network connection lost: unexpected EOF
}

// Example_slogStructuredLogging demonstrates how to use slog for structured logging
//
//nolint:testableexamples
func Example_slogStructuredLogging() {
	err := errors.WithStack(errors.New("missing config"))

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Error("initialization failed", slog.Any("error", err))

	// Example output:
	// {
	//   "time": "2025-06-10T10:01:44.693101023+09:00",
	//   "level": "ERROR",
	//   "msg": "initialization failed",
	//   "error": {
	//     "message": "missing config",
	//     "kind": "*errors.Error",
	//     "stack": [
	//       "loadConfig at config.go:42",
	//       "main at main.go:10"
	//     ]
	//   }
	// }
}

// ExampleRecoverError demonstrates error creation from panic recovery.
//
//nolint:testableexamples
func ExampleRecoverError() {
	safeFunc := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = errors.RecoverError(fmt.Sprintf("database operation failed: %v", r))
			}
		}()

		// This will panic
		panic("connection timeout")
	}

	err := safeFunc()
	fmt.Printf("%+v\n", err)

	// Example Output:
	// database operation failed: connection timeout
	// runtime.gopanic
	// 	/usr/local/go/src/runtime/panic.go:770
	// github.com/stackprune/errors_test.ExampleRecoverError.func1
	// 	/app/example_test.go:208
	// github.com/stackprune/errors_test.ExampleRecoverError
	// 	/app/example_test.go:211
}
