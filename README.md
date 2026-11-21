# stackprune/errors

A modern, lightweight error handling library for Go — focused on:

* ✅ Clean, single-stack trace errors (no duplication)
* ✅ Full compatibility with Go's `errors` package
* ✅ Seamless structured logging via `slog`
* ✅ Panic recovery helpers with preserved stack
* ✅ No external dependencies

---

## Why stackprune/errors?

Go's standard `errors` package does not capture stack traces.
Capturing a new stack on every wrap can produce redundant, noisy stack dumps.

**stackprune/errors captures the stack once at the origin and preserves it across all wraps.**
No redundant stacks. Clean, minimal, highly debuggable output.

---

## Installation

```bash
go get github.com/stackprune/errors/v2
```

### Upgrading from v1

Update imports from `github.com/stackprune/errors` to
`github.com/stackprune/errors/v2` and add the v2 module to your dependencies.
The package name remains `errors`, so calls such as `errors.Wrap` and
`errors.Join` keep the same names. Structured log readers must also adopt the
v2 field layout described below.

---

## Quick Start

```go
import "github.com/stackprune/errors/v2"

// Create error with stack trace
err := errors.New("failed to connect")

// Wrap with additional context (preserves stack)
err = errors.Wrap(err, "while starting server")

// Format errors
err = errors.Errorf("user %d not found", userID)

// Add stack to 3rd party errors
err = errors.WithStack(io.EOF)

// Print with full stack trace
fmt.Printf("%+v\n", err)
```

---

## Key API Summary

| Function           | Purpose                                    |
| ------------------ | ------------------------------------------ |
| `New()`            | Create error, capture stack once           |
| `Wrap()`           | Add context, preserve existing stack                  |
| `WithStack()`      | Attach stack if missing                    |
| `Errorf()`         | Formatted error with stack                 |
| `Join()`           | Combine multiple errors                    |
| `RecoverError()`   | Build error from panic recovery            |
| `NewWithCallers()` | Build error from external program counters |

Full API: [Go Reference](https://pkg.go.dev/github.com/stackprune/errors/v2)

---

## Panic Recovery

Capture panics with full stack trace easily:

```go
func safeOperation() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.RecoverError(fmt.Sprintf("panic: %v", r))
		}
	}()
	panic("something failed")
}
```

For advanced control (external call stacks):

```go
func advancedRecovery() (err error) {
	defer func() {
		if r := recover(); r != nil {
			var pcs [32]uintptr
			n := runtime.Callers(0, pcs[:])
			err = errors.NewWithCallers(fmt.Sprint(r), pcs[:n])
		}
	}()
	panic("something failed")
}
```

---

## Multiple Error Handling

```go
var errs []error

if err := op1(); err != nil {
	errs = append(errs, errors.Wrap(err, "op1 failed"))
}
if err := op2(); err != nil {
	errs = append(errs, errors.Wrap(err, "op2 failed"))
}

return errors.Join(errs...)
```

---

## Structured Logging (slog)

Save this complete example as `main.go` and run `go run -trimpath main.go`:

```go
package main

import (
	"log/slog"
	"os"

	"github.com/stackprune/errors/v2"
)

func connectDB() error {
	return errors.New("database failure")
}

func main() {
	err := connectDB()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Error("operation failed", slog.Any("error", err))
}
```

Example output, pretty-printed for readability. The timestamp varies;
`-trimpath` makes the source paths relative:

```json
{
  "time": "2026-11-21T15:57:45.301511647+09:00",
  "level": "ERROR",
  "msg": "operation failed",
  "error": {
    "message": "database failure",
    "kind": "*errors.Error",
    "stack": [
      "main.connectDB at ./main.go:11",
      "main.main at ./main.go:15"
    ],
    "children": []
  }
}
```

### Joined and wrapped errors

Both `Error` and `JoinError` use the same top-level fields:

| Field | Contents |
| ----- | -------- |
| `message` | The complete error message, including wrapping context and joined messages. |
| `kind` | The error type after following single-error unwraps, stopping at a leaf or a multi-error value. |
| `stack` | The first non-empty stack found in depth-first, left-to-right order, or `[]` if none is available. |
| `children` | Child errors from a multi-error unwrap, or `[]` for a leaf. |

`Wrap` combines its context with the message without adding a JSON level.
`Join` preserves its grouping in the `children` array, including nested joins.
An empty `children` array marks a leaf; the enclosing value still describes an error.
When contained in an `Error` or `JoinError`, standard single-error wrappers and
`Unwrap() []error` implementations, including the standard library's `errors.Join`,
follow the same rules. Nil children are ignored.

Using the same imports and `connectDB` helper, replace `main` with:

```go
func main() {
	dbErr := connectDB()
	cacheErr := errors.New("cache failure")
	err := errors.Wrap(
		errors.Join(errors.Wrap(dbErr, "connect"), cacheErr),
		"startup",
	)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Error("operation failed", slog.Any("error", err))
}
```

Example output from `go run -trimpath main.go`:

```json
{
  "time": "2026-11-21T15:57:45.284710971+09:00",
  "level": "ERROR",
  "msg": "operation failed",
  "error": {
    "message": "startup: connect: database failure\ncache failure",
    "kind": "*errors.JoinError",
    "stack": [
      "main.connectDB at ./main.go:11",
      "main.main at ./main.go:15"
    ],
    "children": [
      {
        "children": [],
        "kind": "*errors.Error",
        "message": "connect: database failure"
      },
      {
        "children": [],
        "kind": "*errors.Error",
        "message": "cache failure"
      }
    ]
  }
}
```

Child entries contain only `message`, `kind`, and `children`; the selected stack
appears only at the top level. Every original error and its stack remain intact
in memory. Error messages, `%+v` formatting, `Is`, `As`, and unwrapping are unaffected.

This is a breaking change to structured log output: the former `errors` field
is now named `children` at every level. Read the representative stack from
`error.stack` instead of `error.errors[i].stack`. Leaf entries now
include `children: []`, and wrapping a join preserves its child array. The same
structure applies to `slog.TextHandler` through grouped attributes.

### Customizing stack output

Set the format before logging an error:

```go
errors.SetLogOptions(errors.LogOptions{
	StackFormat: errors.StackFormatObjectArray,
})
```

For the same captured stack as the example above, the `stack` field
(shown on its own) becomes:

```json
{
  "stack": [
    {
      "file": "./main.go",
      "function": "main.connectDB",
      "line": 11
    },
    {
      "file": "./main.go",
      "function": "main.main",
      "line": 15
    }
  ]
}
```

`MessageKey`, `KindKey`, and `StackKey` customize the corresponding field names.
The message and kind keys also apply to child entries. Use distinct names and
avoid the reserved `children` key.

---

## Debug Output

Use `%+v` to print the complete message and a stack trace. This example
adds context with `Wrap` while preserving the stack captured in `connectDB`:

```go
package main

import (
	"fmt"

	"github.com/stackprune/errors/v2"
)

func connectDB() error {
	return errors.New("database failure")
}

func main() {
	err := errors.Wrap(connectDB(), "startup")
	fmt.Printf("%+v\n", err)
}
```

Output from `go run -trimpath main.go`:

```text
startup: database failure
main.connectDB
	./main.go:10
main.main
	./main.go:14
```

---

## Requirements

```txt
Go 1.26+
Zero external dependencies
```

---

## License

MIT License
