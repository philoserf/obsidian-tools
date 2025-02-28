# Obsidian Tools

A collection of tools for Obsidian plugin development and validation.

## Contents

- [Validate Plugin Manifest](#validate-plugin-manifest)
- [Development](#development)

## Validate Plugin Manifest

A tool to check an Obsidian plugin manifest against community rules as described in the [Validate Plugin Entry workflow](https://github.com/obsidianmd/obsidian-releases/blob/master/.github/workflows/validate-plugin-entry.yml) of the obsidianmd/obsidian-releases project.

### Installation

```bash
# Build from source
go build -o obsidian-validate-plugin-manifest ./validate-plugin-manifest

# Or use Task (recommended)
task build

# Run from source
go run ./validate-plugin-manifest/main.go
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

Run the unit tests with:

```bash
# Using Go directly
go test ./validate-plugin-manifest

# Or using Task (recommended)
task test
```

For more detailed test output:

```bash
# Using Go directly
go test -v ./validate-plugin-manifest

# Or using Task
task test -- -v
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

This project uses [Task](https://taskfile.dev/) as a build tool. You can install it with:

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

Common commands:

```bash
# Build the project
task build

# Run all tests
task test

# Format code
task fmt

# Run linters
task lint

# Install required tools
task tools:install
```

Run `task --list-all` to see all available tasks. See [CLAUDE.md](CLAUDE.md) for more details on development guidelines.
