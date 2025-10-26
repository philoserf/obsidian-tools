# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A collection of tools for Obsidian plugin development and validation. Currently contains a single tool: `validate-plugin-manifest`, which checks Obsidian plugin manifest.json files against community rules from the obsidianmd/obsidian-releases project.

## Build System

This project uses [Task](https://taskfile.dev/) as its build tool. All common operations are defined in `Taskfile.yml`.

### Essential Commands

```bash
# Build the validator tool
task build

# Run all tests
task test
task test:cover  # With coverage
task test:race   # With race detection

# Code quality
task lint        # Run all linters (Go + Markdown)
task format      # Format all code
task fix         # Auto-fix linting and formatting issues

# Run the validator
task run         # Default validation
task run:json    # JSON output
```

### Testing

When running tests directly with Go, pass extra flags after `--`:

```bash
# Verbose test output
task test -- -v

# Run tests in a specific package
go test -v ./validate-plugin-manifest
```

## Code Architecture

### Tool Structure

The codebase follows a simple package-per-tool pattern:

- `validate-plugin-manifest/` - Self-contained tool package
  - `main.go` - CLI entry point, manifest validation logic
  - `output.go` - Output formatting (text and JSON)
  - `*_test.go` - Unit tests

Each tool is a standalone Go package with its own `main()` function.

### Validation Architecture

The validator implements a modular validation pattern:

1. **Manifest struct** - Defines the expected JSON structure
2. **ValidationResult** - Accumulates errors and warnings
3. **ValidateManifest()** - Orchestrates all validation checks
4. **Individual validators** - Focused functions like `validateID()`, `validateName()`, etc.

Each validator function receives the manifest and a shared `ValidationResult` to accumulate issues. This makes it easy to add new validation rules.

### Output Formatting

Output is abstracted through the `OutputFormat` type with two implementations:

- `TextOutput` - Human-readable with emojis
- `JSONOutput` - Machine-readable for CI/CD

The `PrintResultsTo()` function accepts an `io.Writer`, making output testable.

## Development Dependencies

Managed via Brewfile:

- `go` - Go compiler
- `golangci-lint` - Go linter
- `gofumpt` - Strict Go formatter
- `markdownlint-cli` - Markdown linter
- `prettier` - Markdown formatter
- `go-task` - Task runner

First-time setup: `task setup` (runs `brew bundle`)

## Adding New Tools

When adding new tools to this repository:

1. Create a new directory for the tool (e.g., `new-tool/`)
2. Implement as a standalone Go package with `package main`
3. Add build target to `Taskfile.yml` under `build:all`
4. Update `BUILD_DIR` and `BINARY_NAME` vars or add new tool-specific vars
5. Add tests following the pattern in `validate-plugin-manifest/`

## Code Standards

- Go version: 1.24 (specified in go.mod)
- All Go code must pass `golangci-lint` and be formatted with `gofumpt`
- All Markdown must pass `markdownlint` and be formatted with `prettier`
- Pre-compiled regexes are defined as package-level `var` for performance
- Functions are organized by responsibility with clear comments
- Error messages should be user-friendly and actionable
