# Obsidian Tools

A collection of tools for Obsidian plugin development and validation.

## Contents

- [Validate Plugin Manifest](#validate-plugin-manifest)
- [Development](#development)

## Project documentation

- [Walkthrough](walkthrough.md) — linear, executable tour of the code
- [Theory](theory.md) — the conceptual model a maintainer needs to hold in mind

## Validate Plugin Manifest

A tool to check an Obsidian plugin manifest against community rules as described in the [Validate Plugin Entry workflow](https://github.com/obsidianmd/obsidian-releases/blob/master/.github/workflows/validate-plugin-entry.yml) of the obsidianmd/obsidian-releases project.

### Installation

The easiest way to install is using Homebrew:

```bash
# Clone the repository
git clone https://github.com/yourusername/obsidian-tools.git
cd obsidian-tools

# Install dependencies and build tools
brew bundle

# Build the tool
task build
```

You can also build directly with Go if you prefer:

```bash
go build -o obsidian-validate-plugin-manifest ./validate-plugin-manifest
```

### Usage

```bash
# Simple validation with default manifest.json
./obsidian-validate-plugin-manifest

# Validate a specific manifest file
./obsidian-validate-plugin-manifest --manifest path/to/manifest.json

# Output in JSON format (useful for CI/CD)
./obsidian-validate-plugin-manifest --manifest path/to/manifest.json --json

# Suppress informational output
./obsidian-validate-plugin-manifest --manifest path/to/manifest.json --quiet

# Show version information
./obsidian-validate-plugin-manifest --version
```

### Validation Rules

The validator checks that manifest files comply with the Obsidian community plugin guidelines:

- Plugin ID and name must not contain "obsidian" or end with "plugin"
- Description must not contain "obsidian" or phrases like "this plugin"
- Description should be under 250 characters
- Version and minAppVersion follow proper format (numbers and dots only)
- URLs don't point to the Obsidian website
- Email addresses are discouraged in the author field

### Testing

```bash
task test            # or: go test ./validate-plugin-manifest
task test -- -v      # verbose
task test -- -cover  # with coverage
```

### Example Output

For successful validation:

```text
📝 Validating manifest for plugin: Example Plugin

✅ Manifest validation passed!
```

For validation with errors:

```text
📝 Validating manifest for plugin: Example Plugin

❌ Errors:
  • Plugin ID should not contain the word 'obsidian'
  • Plugin ID should not end with 'plugin'
  • Description should not contain the word 'Obsidian'
  • Description should be under 250 characters (currently 274)

⚠️  Warnings:
  • Email addresses are discouraged in the author field

❌ Validation failed with 4 error(s) and 1 warning(s)
```

## Development

This project uses [Homebrew](https://brew.sh) for dev dependencies and [Task](https://taskfile.dev/) as a build tool. Initial setup: `brew bundle`.

```bash
task build           # Build the binary
task test            # Run tests (pass flags via `--`, e.g. `task test -- -cover`)
task lint            # Run linters
task fix             # Auto-fix lint and formatting (Go + Markdown)
task run             # Run the validator (pass flags via `--`, e.g. `task run -- --json`)
task ci              # lint + test
```

Run `task` (or `task --list-all`) for the full list.
