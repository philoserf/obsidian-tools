# Obsidian Tools Development Guide

## Project Setup

This project uses [Homebrew](https://brew.sh) to manage dependencies and [Task](https://taskfile.dev/) as a build tool.

### Initial Setup

To set up your development environment:

```bash
# Clone the repository
git clone https://github.com/yourusername/obsidian-tools.git
cd obsidian-tools

# Install all dependencies using Homebrew
brew bundle

# Or use the convenience task (if you already have Task installed)
task setup
```

This command will install all required dependencies defined in the Brewfile, including:

- Go
- Task
- golangci-lint
- markdownlint-cli
- prettier
- Other development tools

### Common Task Commands

- `task setup` - Setup development environment
- `task verify` - Verify development tools are correctly installed
- `task build` - Build the project
- `task test` - Run all tests
- `task format` - Format all code
- `task lint` - Run all linters
- `task fix` - Fix all linting and formatting issues
- `task quality` - Run all linters and formatters
- `task ci` - Run all checks (lint, format, test)
- `task pre-commit` - Run checks before committing

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

- Format Go code: `task format:go` or `gofumpt -l -w .`
- Format Markdown: `task format:markdown` or `prettier --write "**/*.md"`
- Lint Go code: `task lint:go` or `golangci-lint run ./...`
- Lint Markdown: `task lint:markdown` or `markdownlint *.md`
- Fix all linting issues: `task lint:fix`

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
