# Obsidian Tools

A collection of tools for Obsidian plugin development and validation.

## Contents

- [Validate Plugin Manifest](#validate-plugin-manifest)
- [Development](#development)

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

This project uses [Homebrew](https://brew.sh) for dependency management and [Task](https://taskfile.dev/) as a build tool.

### Setting Up Development Environment

```bash
# Initial setup - installs all dependencies from Brewfile
task setup

# Verify your environment is correctly set up
task verify
```

### Common Task Commands

```bash
# Build the project
task build

# Run all tests
task test
task test:cover  # Run with coverage report
task test:race   # Run with race detection

# Format code
task format

# Run linters
task lint

# Fix linting and formatting issues
task fix

# Run the validator
task run
task run:json    # With JSON output
```

### Task Categories

The Taskfile is organized into logical sections:

1. **Setup Tasks**: `setup`, `verify`
2. **Build Tasks**: `build`, `build:all`, `clean`
3. **Quality Tasks**: `quality`, `fix`, `format`, `lint`
4. **Test Tasks**: `test`, `test:cover`, `test:race`
5. **Run Tasks**: `run`, `run:json`
6. **Maintenance Tasks**: `update:deps`
7. **CI/CD Tasks**: `ci`, `pre-commit`

Run `task --list-all` to see all available tasks. See [copilot instructions](.github/copilot-instructions.md) for more detailed development guidelines.
