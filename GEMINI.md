# Proteus

## Project Overview

Proteus is a Go library for managing application configuration. It allows developers to define the configuration of an application in a struct and load it from different sources like environment variables and command-line flags. The library also supports configuration updates.

The project uses `golangci-lint` for linting and has a comprehensive test suite.

## Building and Running

### Dependencies

The project has one dependency: `golang.org/x/exp`.

### Linting

To run the linter, use the following command:

```bash
make check
```

This will run `golangci-lint` on the entire codebase.

### Testing

To run the tests, use the following command:

```bash
make test
```

This will run all the tests with the race detector enabled.

### Test Coverage

To generate a test coverage report, use the following command:

```bash
make cover
```

This will generate a coverage report and open it in your browser.

## Development Conventions

The project follows standard Go conventions. All code is formatted using `gofmt`. The project uses a `Makefile` to automate common tasks like linting, testing, and generating coverage reports.
