# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-06-16

### Added
- `NewWithCallers` function for creating errors with custom stack traces from program counters, particularly useful for panic/recover scenarios where you need to preserve the original stack context.
- `RecoverError` helper function for convenient error creation from panic recovery values with proper stack trace capture.
- `ProgramCounters()` method on `Error` type for accessing raw program counters (primarily for testing and advanced debugging).

### Improved
- Enhanced panic recovery capabilities with dedicated helper functions for common patterns.
- Improved stack trace filtering to show only relevant frames (`runtime.goexit`, `runtime.main` instead of all `runtime.*` functions).
- Better integration with defer/recover patterns commonly used in Go applications.

### Documentation
- Major rewrite of README.md for improved usability and clarity.


## [1.0.2] - 2025-06-12

### Improved
- Enhanced `SetLogOptions` function to automatically apply default values for empty keys, improving usability.
- Reorganized README.md structure for better readability:
  - Moved structured logging (`slog`) section after basic usage examples
  - Improved section titles and flow
  - Fixed documentation inaccuracies in slog usage examples

### Fixed
- Corrected README.md documentation errors:
  - Fixed non-existent function name (`SetStackFormatter` → `SetLogOptions`)
  - Fixed non-existent constant name (`StackAsObjects` → `StackFormatObjectArray`)
  - Fixed JSON field name in examples (`"func"` → `"function"`)


## [1.0.1] - 2025-06-10

### Documentation
- Added `slog` usage examples to `README.md`.
- Expanded `doc.go` with structured logging and stack trace formatting details.


## [1.0.0] - 2025-06-09

Initial stable release with structured logging support.

### Features
- Implemented `slog.LogValuer` for `*Error`, enabling structured logging of error messages and stack traces.
- Supports configurable stack trace formats:
  - `string array` (default)
  - `object array` (structured)


## [v0.1.1] - 2025-06-09

### Changed
- Stack trace output now stops after encountering any function starting with `runtime.`.
  This reduces noise from entries such as `runtime.goexit` or `runtime.goext`, making traces cleaner and easier to read.


## [v0.1.0] - 2025-06-08

### Added
- First release.
