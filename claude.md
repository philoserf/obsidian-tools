# Obsidian Tools Development Guide

## Task Runner

This project uses [Task](https://taskfile.dev/) as a build tool. You can install it with:

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

### Common Task Commands

- `task build` - Build the project
- `task test` - Run all tests
- `task fmt` - Format all code
- `task lint` - Run all linters
- `task lint:fix` - Run all linters with autofix
- `task tools:install` - Install required development tools

Run `task --list-all` to see all available tasks.

## Commands

### Build and Run

- Build: `go build -o obsidian-validate-plugin-manifest ./validate-plugin-manifest`
- Run: `go run ./validate-plugin-manifest/main.go`
- Run with custom manifest: `go run ./validate-plugin-manifest/main.go --manifest path/to/manifest.json`
- JSON output: `go run ./validate-plugin-manifest/main.go --json`
- Quiet mode: `go run ./validate-plugin-manifest/main.go --quiet`
- Show version: `go run ./validate-plugin-manifest/main.go --version`

### Test

- Run all tests: `go test ./validate-plugin-manifest`
- Test with verbose output: `go test -v ./validate-plugin-manifest`
- Test coverage: `go test -cover ./validate-plugin-manifest`

### Format and Lint

- Format Go code: `gofumpt -l -w .`
- Format Markdown: `prettier --write "**/*.md"`
- Lint Go code: `golangci-lint run ./...`
- Lint Markdown: `markdownlint *.md`

## Code Style Guidelines

- **Imports**: Group standard library imports first, followed by third-party packages
- **Error Handling**: Use explicit error checking with appropriate logging to stderr
- **Naming**:
  - Use camelCase for variables and functions
  - Use PascalCase for exported types and functions
  - Use snake_case for test file names
- **Formatting**: Follow standard Go formatting guidelines (use gofumpt, not gofmt)
- **Documentation**: Include comments for all exported functions, types, and constants
- **Validation**: Use AddErrorf/AddWarningf methods for collecting validation results
- **Testing**: Write tests for all exported functions and validation rules
  - Always use t.Parallel() in tests when possible
  - Use descriptive variable names (avoid short names like tt)
  - Break down large test functions into smaller helper functions
