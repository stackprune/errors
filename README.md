# stackprune/errors

`stackprune/errors` is a lightweight, modern error handling library for Go, designed as a clean alternative to `pkg/errors`.

It captures **one stack trace** at the point where the error is created and preserves it through all wrapping, avoiding redundant trace output. This makes the %+v output readable and focused — exactly what you need for debugging.

## Why Use stackprune/errors?

- ✅ Clean, single-source stack trace (no duplication)
- ✅ Minimalistic API inspired by `pkg/errors`
- ✅ `%+v` formatting shows clear trace
- ✅ Fully compatible with Go 1.13+ error wrapping (`errors.Is`, `errors.As`, `errors.Unwrap`)
- ✅ No `WithMessage` clutter — just `Wrap()`
- ✅ Lightweight and dependency-free

---

## Why not `pkg/errors`?

`pkg/errors` captures a new stack trace every time you call `Wrap()`. This results in bloated, duplicated stack outputs:

```go
err := errors.New("root cause")        // <- captures stack
err = errors.Wrap(err, "level 2")      // <- new stack
err = errors.Wrap(err, "level 3")      // <- yet another stack
```

With `stackprune/errors`, stack is captured **only once**:

```go
err := errors.New("root cause")        // captures stack here
err = errors.Wrap(err, "level 2")      // adds message, no new stack
err = errors.Wrap(err, "level 3")      // same original stack
```

---

## Installation

```bash
go get github.com/stackprune/errors
```

---

## Core API Comparison

| Function      | stackprune/errors                           | pkg/errors                 |
|---------------|---------------------------------------------|----------------------------|
| `New()`       | Captures stack once                         | Captures stack             |
| `Wrap()`      | Adds message, no new stack                  | Adds message + stack       |
| `WithStack()` | Adds stack if missing                       | Always adds stack          |
| `Errorf()`    | Formats message and captures stack          | Formats and captures stack |
| `Join()`      | Uses Go 1.20+ `errors.Join` (with fallback) | Not available              |

---

## Usage

Basic error creation and wrapping:

```go
import "github.com/stackprune/errors"

func loadConfig() error {
    return errors.New("missing config file")
}

func runApp() error {
    err := loadConfig()
    return errors.Wrap(err, "failed to start app")
}
```

Formatted error messages:

```go
if err := doSomething(); err != nil {
    return errors.Errorf("operation failed: %w", err)
}
```

Adding a stack trace to third-party errors:

```go
if err := thirdPartyFunc(); err != nil {
    return errors.WithStack(err)
}
```

Combining multiple errors:

```go
var errs []error

if err := op1(); err != nil {
    errs = append(errs, err)
}
if err := op2(); err != nil {
    errs = append(errs, err)
}

return errors.Join(errs...)
```

Pretty-printing with stack trace:

```go
fmt.Printf("%+v\n", err) // shows full stack from the point where error was created
```

### Notes

* This library requires **Go 1.22+**. Older versions are not supported or tested.
* `Join` provides the same semantics as Go 1.20's `errors.Join`, with internal fallback for compatibility.
* Stack traces are only captured once — even when wrapping multiple times — keeping logs concise.

---

## Example Output

```go
if err := handler_createUser(); err != nil {
    fmt.Printf("%+v\n", err)
}
```

Output:

```
user creation failed: failed to insert user into database
github.com/stackprune/errors_test.repository_insertUser
    /app/main.go:133
github.com/stackprune/errors_test.usecase_createUser
    /app/main.go:125
github.com/stackprune/errors_test.handler_createUser
    /app/main.go:117
github.com/stackprune/errors_test.ExampleError_Format_stackTrace
    /app/main.go:139
```

✔ Stack trace shown once  
✔ No duplication  
✔ Easy to follow

---

## API Reference

See [Go Reference](https://pkg.go.dev/github.com/stackprune/errors) for full API documentation.

---

## License

MIT License
