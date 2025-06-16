# stackprune/errors

A modern, lightweight error handling library for Go — focused on:

* ✅ Clean, single-stack trace errors (no duplication)
* ✅ Full compatibility with Go's `errors` package
* ✅ Seamless structured logging via `slog`
* ✅ Panic recovery helpers with preserved stack
* ✅ No external dependencies

---

## Why stackprune/errors?

Go's standard `errors` and popular `pkg/errors` often capture new stack traces on each wrap, leading to redundant, noisy stack dumps.

**stackprune/errors captures the stack once at the origin and preserves it across all wraps.**
No redundant stacks. Clean, minimal, highly debuggable output.

---

## Installation

```bash
go get github.com/stackprune/errors
```

---

## Quick Start

```go
import "github.com/stackprune/errors"

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
| `Wrap()`           | Add context, no new stack                  |
| `WithStack()`      | Attach stack if missing                    |
| `Errorf()`         | Formatted error with stack                 |
| `Join()`           | Combine multiple errors                    |
| `RecoverError()`   | Build error from panic recovery            |
| `NewWithCallers()` | Build error from external program counters |

Full API: [Go Reference](https://pkg.go.dev/github.com/stackprune/errors)

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

Seamless integration with Go's `slog`:

```go
err := errors.New("database failure")
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Error("operation failed", slog.Any("error", err))
```

**Output example:**

```json
{
  "level": "ERROR",
  "msg": "operation failed",
  "error": {
    "message": "database failure",
    "kind": "*errors.Error",
    "stack": [
      "connectDB at db.go:42",
      "main at main.go:10"
    ]
  }
}
```

### Customizing stack output

```go
errors.SetLogOptions(errors.LogOptions{
	StackFormat: errors.StackFormatObjectArray,
})
```

Produces:

```json
"stack": [
  {"file": "db.go", "function": "connectDB", "line": 42},
  {"file": "main.go", "function": "main", "line": 10}
]
```

---

## Debug Output

Use `%+v` to print full error with stack trace:

```go
fmt.Printf("%+v\n", err)
```

Example output:

```
user creation failed: failed to insert
main.repositoryInsertUser
    /app/main.go:133
main.usecaseCreateUser
    /app/main.go:125
main.handleCreateUser
    /app/main.go:117
main.main
    /app/main.go:139
```

---

## Requirements

```txt
Go 1.22+
Zero external dependencies
```

---

## License

MIT License
