# AGENTS.md

## Project Overview

A collection of tools for Obsidian plugin development and validation. Currently contains one tool: `validate-plugin-manifest`, which checks Obsidian plugin manifest.json files against community rules from the obsidianmd/obsidian-releases project.

## Build & Development Commands

This project uses [Task](https://taskfile.dev/) as its build tool. First-time setup: `task setup` (runs `brew bundle`).

```bash
task build           # Build the validator binary
task test            # Run all tests (verbose)
task test:cover      # Tests with coverage report
task test:race       # Tests with race detection
task lint            # Run all linters (Go + Markdown)
task fix             # Auto-fix linting and formatting issues
task ci              # Full CI pipeline (quality + test)
task run             # Run the validator
task run:json        # Run with JSON output
```

Pass extra flags via `--`: `task test -- -run TestValidateManifest`

Direct Go commands also work: `go test -v ./validate-plugin-manifest`

## Code Standards

- **Go 1.25** — pure standard library, no external dependencies
- **Formatter:** `gofumpt` (stricter than gofmt)
- **Linter:** `golangci-lint`
- **Markdown:** `prettier` for formatting
- Pre-compiled regexes as package-level `var` for performance

## Architecture

### Package-per-tool pattern

Each tool is a standalone `package main` directory. The current tool lives at `validate-plugin-manifest/`:

- `main.go` — CLI entry point, `Manifest` struct, `ValidationResult` accumulator, individual validator functions (`validateID`, `validateName`, etc.)
- `output.go` — Output formatting abstraction via `OutputFormat` type (`TextOutput`, `JSONOutput`). `PrintResultsTo()` accepts `io.Writer` for testability.
- `*_test.go` — Table-driven, parallel tests using a `TestCase` struct and `checkValidationResult()` helper.

### Validation pattern

`ValidateManifest()` orchestrates individual validators. Each validator receives the manifest and a shared `ValidationResult` pointer to accumulate errors and warnings. Adding a new rule means writing a focused function and calling it from `ValidateManifest()`.

## Adding New Tools

1. Create a new directory (e.g., `new-tool/`)
2. Implement as `package main` with its own `main()` function
3. Add build targets in `Taskfile.yml` under `build:all`
4. Follow existing test patterns (table-driven, parallel, helper assertions)
