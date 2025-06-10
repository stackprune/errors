# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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