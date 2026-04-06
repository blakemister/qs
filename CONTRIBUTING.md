# Contributing to qs

Thanks for your interest in contributing!

## Prerequisites

- **Go 1.24+**
- **Windows 10/11**
- **Windows Terminal**

## Dev Setup

```bash
git clone https://github.com/bcmister/qs.git
cd qs
go mod tidy
go build -o qs.exe .
go test ./...
```

## Code Style

- Standard Go formatting (`gofmt`)
- Lint with `go vet ./...` before submitting
- Follow existing patterns in the codebase
- TUI views use the Bubble Tea Elm architecture (`Init`, `Update`, `View`)
- Styles are centralized in `internal/tui/styles.go` — use existing style variables

## Making Changes

1. Fork the repo and create a branch from `main`
2. Make your changes in small, focused commits
3. Add tests for new functionality
4. Run the full test suite: `go test ./...`
5. Run the linter: `go vet ./...`
6. Verify it builds: `go build -o qs.exe .`
7. Open a pull request against `main`

## Build Flags Warning

**Do NOT add `-s -w` to ldflags.** The binary must not be stripped (WDAC policy). The `-X` flag for version injection is safe and is the only ldflags modifier used.

## Testing

```bash
go test ./...          # Run all tests
go test ./internal/config/...  # Run config tests only
go test -v ./...       # Verbose output
```

Add tests for any new exported functions. Tests live alongside their source files (`*_test.go`).

## Project Structure

```
main.go                  Entry point
internal/
  cmd/                   CLI commands (Cobra)
  config/                Config loading, migration, accounts
  launcher/              Win32 window spawning + positioning
  monitor/               Win32 monitor detection
  tui/                   All Bubble Tea TUI views
```

## Questions?

Open an issue or start a discussion on the repo.
