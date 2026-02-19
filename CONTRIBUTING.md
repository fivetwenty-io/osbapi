# Contributing to osbapi

Thank you for your interest in contributing to the osbapi Go library.

## Prerequisites

- Go 1.22 or later

- `make` (for running project targets)

## Getting Started

```bash
git clone https://github.com/fivetwenty-io/osbapi.git
cd osbapi
go mod download
```

## Running Tests

```bash
# Run all unit tests
make test

# Run tests with race condition detection
make test-race

# Run integration tests (starts an in-memory broker)
make test-integration

# Generate a coverage report
make coverage
```

## Code Quality

```bash
# Format code and run go vet
make lint

# Run fmt, vet, staticcheck, and tests
make check

# Run all security scanners (govulncheck, gosec, trivy)
make security
```

## Project Structure

```
pkg/
  osbapi/       Shared types, interfaces, and errors
  broker/       Server-side entry point (broker.NewHandler)
  osbclient/    Client-side entry point (osbclient.New)
internal/
  server/       HTTP handler implementation
  client/       HTTP client implementation
  constants/    Shared constants
examples/
  broker/       Example service broker
  platform/     Example platform client
test/
  integration/  End-to-end integration tests
```

## Pull Request Process

1. Fork the repository and create a feature branch.

2. Write tests for any new functionality.

3. Ensure all checks pass: `make check`.

4. Keep commits atomic with clear, imperative-mood messages.

5. Open a pull request against the `main` branch with a description of the change.
