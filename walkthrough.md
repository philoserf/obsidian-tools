# obsidian-tools Walkthrough

*2026-05-26T16:14:43Z by Showboat 0.6.1*
<!-- showboat-id: 3fef862a-9994-4849-9419-1bce154d9b1b -->

## Overview

`obsidian-tools` is a workspace for Obsidian-related developer utilities. Today it contains a single tool: **validate-plugin-manifest**, a CLI that checks a plugin's `manifest.json` against the rules the Obsidian community publishes in `obsidianmd/obsidian-releases`.

Key facts:

- **Language:** Go 1.25, standard library only — no third-party imports.
- **Layout:** Each tool is its own `package main` directory. Adding a tool means adding a directory.
- **Entry point:** `validate-plugin-manifest/main.go`.
- **Outputs:** Human-readable text (default) or machine-readable JSON (`--json`).
- **Exit code:** Non-zero when validation reports any errors; warnings alone are not fatal.

The validator runs a fixed set of focused validator functions, each of which appends to a shared `ValidationResult`. Adding a new rule is a matter of writing a function and calling it from `ValidateManifest`.

## Architecture

The repository root holds workspace-level configuration (`Taskfile.yml`, `Brewfile`, `go.mod`, `README.md`). Each tool lives in its own subdirectory as a standalone `package main`.

```bash
ls -1 /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/
```

```output
main.go
output_test.go
output.go
validator_test.go
```

Two source files split responsibility cleanly:

- `main.go` — CLI entry point, the `Manifest` data model, the `ValidationResult` accumulator, and all rule-specific validator functions.
- `output.go` — Output formatting abstraction (`OutputFormat`, `PrintResultsTo`), with both text and JSON renderers writing to an `io.Writer` for testability.

The module declares no third-party dependencies.

```bash
cat /Users/markayers/source/philoserf/obsidian-tools/go.mod
```

```output
module obsidian-validate-plugin-manifest

go 1.25
```

## Entry point: `main()`

The CLI uses the standard `flag` package. It accepts `--manifest` (path, default `manifest.json`), `--version`, `--json`, and `--quiet`.

```bash
sed -n '203,221p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
func main() {
	var (
		manifestPath string
		showVersion  bool
		jsonOutput   bool
		quiet        bool
	)

	flag.StringVar(&manifestPath, "manifest", "manifest.json", "Path to manifest.json file")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	flag.BoolVar(
		&quiet,
		"quiet",
		false,
		"Suppress informational output (only shows errors/warnings)",
	)
	flag.Parse()

```

After parsing flags, `main()` short-circuits on `--version`, resolves the manifest path to absolute, reads and JSON-decodes the file, then dispatches to validation and output. Errors during read/parse go to `stderr` and exit `1`.

```bash
sed -n '222,266p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
	if showVersion {
		// Print version information to stdout
		_, _ = fmt.Fprintf(os.Stdout, "Obsidian Plugin Manifest Validator v%s\n", version)

		return
	}

	if !filepath.IsAbs(manifestPath) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		manifestPath = filepath.Join(cwd, manifestPath)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading manifest file '%s': %v\n", manifestPath, err)
		os.Exit(1)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing manifest.json: %v\n", err)
		os.Exit(1)
	}

	// Use JSON output format if requested
	outputFormat := TextOutput
	if jsonOutput {
		outputFormat = JSONOutput
	} else if !quiet {
		// Print status information to stdout
		_, _ = fmt.Fprintf(os.Stdout, "📝 Validating manifest for plugin: %s\n", manifest.Name)
	}

	result := ValidateManifest(&manifest)
	PrintResults(result, outputFormat)

	if !result.IsValid() {
		os.Exit(1)
	}
}
```

## Data model

The `Manifest` struct mirrors the fields Obsidian's plugin manifest defines. `AuthorURL` and `FundingURL` are optional (`,omitempty`).

```bash
sed -n '25,36p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// Manifest represents an Obsidian plugin manifest.json structure.
type Manifest struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Author        string `json:"author"`
	Version       string `json:"version"`
	MinAppVersion string `json:"minAppVersion"`
	IsDesktopOnly bool   `json:"isDesktopOnly"`
	AuthorURL     string `json:"authorUrl,omitempty"`
	FundingURL    string `json:"fundingUrl,omitempty"`
}
```

## The accumulator: `ValidationResult`

Validators don't return errors directly — they push messages onto a shared `ValidationResult`. The struct keeps errors and warnings as separate string slices. `IsValid()` is the single source of truth for "did we pass?" — warnings never flip it to false.

`AddError`/`AddWarning` take a literal string; `AddErrorf`/`AddWarningf` use `fmt.Sprintf` formatting. This pair-API keeps the call sites at validators readable without per-site string building.

```bash
sed -n '38,67p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// ValidationResult stores validation errors and warnings.
type ValidationResult struct {
	Errors   []string
	Warnings []string
}

// AddError adds an error message to the validation result.
func (vr *ValidationResult) AddError(message string) {
	vr.Errors = append(vr.Errors, message)
}

// AddErrorf adds a formatted error message to the validation result.
func (vr *ValidationResult) AddErrorf(format string, args ...any) {
	vr.Errors = append(vr.Errors, fmt.Sprintf(format, args...))
}

// AddWarning adds a warning message to the validation result.
func (vr *ValidationResult) AddWarning(message string) {
	vr.Warnings = append(vr.Warnings, message)
}

// AddWarningf adds a formatted warning message to the validation result.
func (vr *ValidationResult) AddWarningf(format string, args ...any) {
	vr.Warnings = append(vr.Warnings, fmt.Sprintf(format, args...))
}

// IsValid returns true if there are no validation errors.
func (vr *ValidationResult) IsValid() bool {
	return len(vr.Errors) == 0
}
```

## Orchestration: `ValidateManifest`

`ValidateManifest` constructs an empty result and calls each rule function in sequence. Every validator gets a pointer to the same result and appends to it — there is no error short-circuiting between validators, so a manifest with many problems reports them all in one run.

```bash
sed -n '69,85p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// ValidateManifest checks the provided manifest against Obsidian community rules.
func ValidateManifest(manifest *Manifest) ValidationResult {
	result := ValidationResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	validateID(manifest, &result)
	validateName(manifest, &result)
	validateDescription(manifest, &result)
	validateAuthor(manifest, &result)
	validateVersion(manifest, &result)
	validateURLs(manifest, &result)
	validateMinAppVersion(manifest, &result)

	return result
}
```

## Pre-compiled regexes

Three regexes are compiled once at package init: lowercase plugin ID characters, an email pattern, and a numeric-dot version pattern. Hoisting them avoids re-compiling on every call.

```bash
sed -n '18,23p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
var (
	// Pre-compiled regular expressions for better performance.
	idRegex      = regexp.MustCompile(`^[a-z0-9-_]+$`)
	emailRegex   = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	versionRegex = regexp.MustCompile(`^[0-9.]+$`)
)
```

## Rule: `validateID`

The first validator returns early on an empty ID — without that guard the remaining checks would produce noisy follow-on errors. Then three rules: no "obsidian" substring, no "plugin" suffix, and a regex shape check.

Notice each rule appends its own message; no early return between them. A maximally-bad ID like `bad-obsidian-plugin` triggers all three.

```bash
sed -n '87,108p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// validateID checks plugin ID rules.
func validateID(manifest *Manifest, result *ValidationResult) {
	if manifest.ID == "" {
		result.AddError("Plugin ID is required")

		return
	}

	if strings.Contains(strings.ToLower(manifest.ID), "obsidian") {
		result.AddError("Plugin ID should not contain the word 'obsidian'")
	}

	if strings.HasSuffix(strings.ToLower(manifest.ID), "plugin") {
		result.AddError("Plugin ID should not end with 'plugin'")
	}

	if !idRegex.MatchString(manifest.ID) {
		result.AddError(
			"Plugin ID must contain only lowercase alphanumeric characters, dashes, and underscores",
		)
	}
}
```

## Rule: `validateName`

Same shape as `validateID` but with `Name` semantics — empty is fatal, contains "obsidian" is an error, ends with "plugin" is an error. No regex check here (names allow any characters).

```bash
sed -n '110,125p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// validateName checks plugin name rules.
func validateName(manifest *Manifest, result *ValidationResult) {
	if manifest.Name == "" {
		result.AddError("Plugin name is required")

		return
	}

	if strings.Contains(strings.ToLower(manifest.Name), "obsidian") {
		result.AddError("Plugin name should not contain the word 'Obsidian'")
	}

	if strings.HasSuffix(strings.ToLower(manifest.Name), "plugin") {
		result.AddError("Plugin name should not end with 'Plugin'")
	}
}
```

## Rule: `validateDescription`

Description rules add two new flavors: a **warning** (not error) for "this plugin" phrasing, and an `AddErrorf` use for the length check, demonstrating the formatted accumulator method.

The `maxDescriptionLength` constant is declared *inside* the function — it is only used here, so colocating it keeps the rule self-contained.

```bash
sed -n '127,149p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// validateDescription checks description rules.
func validateDescription(manifest *Manifest, result *ValidationResult) {
	if manifest.Description == "" {
		result.AddError("Description is required")

		return
	}

	if strings.Contains(strings.ToLower(manifest.Description), "obsidian") {
		result.AddError("Description should not contain the word 'Obsidian'")
	}

	if strings.Contains(strings.ToLower(manifest.Description), "this plugin") {
		result.AddWarning("Avoid phrases like 'this plugin' in the description")
	}

	// maxDescriptionLength defines the maximum allowed length for the plugin description.
	const maxDescriptionLength = 250
	if len(manifest.Description) > maxDescriptionLength {
		result.AddErrorf("Description should be under %d characters (currently %d)",
			maxDescriptionLength, len(manifest.Description))
	}
}
```

## Rule: `validateAuthor`, `validateVersion`, `validateMinAppVersion`

The remaining single-field rules are short. Author requires non-empty and warns if it looks like an email. Version and minAppVersion both require non-empty and must match the numeric-dot regex (so `1.0.0-beta` is rejected).

```bash
sed -n '151,175p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// validateAuthor checks author field rules.
func validateAuthor(manifest *Manifest, result *ValidationResult) {
	if manifest.Author == "" {
		result.AddError("Author is required")

		return
	}

	if emailRegex.MatchString(manifest.Author) {
		result.AddWarning("Email addresses are discouraged in the author field")
	}
}

// validateVersion checks version format rules.
func validateVersion(manifest *Manifest, result *ValidationResult) {
	if manifest.Version == "" {
		result.AddError("Version is required")

		return
	}

	if !versionRegex.MatchString(manifest.Version) {
		result.AddError("Invalid version number format (should only contain numbers and dots)")
	}
}
```

```bash
sed -n '188,199p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// validateMinAppVersion checks minAppVersion rules.
func validateMinAppVersion(manifest *Manifest, result *ValidationResult) {
	if manifest.MinAppVersion == "" {
		result.AddError("MinAppVersion is required")

		return
	}

	if !versionRegex.MatchString(manifest.MinAppVersion) {
		result.AddError("Invalid minAppVersion format (should only contain numbers and dots)")
	}
}
```

## Rule: `validateURLs`

The URL validator is purely targeted at two specific banned URLs — there is no general well-formedness check. Anything else is allowed (and the fields themselves are optional).

```bash
sed -n '177,186p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/main.go
```

```output
// validateURLs checks author and funding URL rules.
func validateURLs(manifest *Manifest, result *ValidationResult) {
	if manifest.AuthorURL == "https://obsidian.md" {
		result.AddError("Author URL should not point to Obsidian website")
	}

	if manifest.FundingURL != "" && manifest.FundingURL == "https://obsidian.md/pricing" {
		result.AddError("Funding URL should not point to Obsidian pricing")
	}
}
```

## Output formatting (`output.go`)

`OutputFormat` is a typed string with two values — `TextOutput` and `JSONOutput`. The `OutputResult` struct is the JSON wire format, with `omitempty` so empty error/warning slices vanish from the output.

```bash
sed -n '10,25p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/output.go
```

```output
// OutputFormat defines the format for validation results.
type OutputFormat string

const (
	// TextOutput is the standard human-readable text output.
	TextOutput OutputFormat = "text"
	// JSONOutput produces machine-readable JSON output.
	JSONOutput OutputFormat = "json"
)

// OutputResult represents the structure used for JSON output.
type OutputResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
```

The print API is two layers: `PrintResultsTo(io.Writer, ...)` does the work; `PrintResults(...)` is a thin wrapper that targets `os.Stdout`. Tests use the `Writer` form against a `bytes.Buffer` — `main()` uses the stdout wrapper.

Format dispatch is a `switch` on `OutputFormat`. An unknown value falls through to the text renderer.

```bash
sed -n '27,42p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/output.go
```

```output
// PrintResultsTo displays validation results to the given writer in the specified format.
func PrintResultsTo(writer io.Writer, result ValidationResult, format OutputFormat) {
	switch format {
	case JSONOutput:
		printJSONResults(writer, result)
	case TextOutput:
		printTextResults(writer, result)
	default:
		printTextResults(writer, result)
	}
}

// PrintResults displays validation results to stdout in the specified format.
func PrintResults(result ValidationResult, format OutputFormat) {
	PrintResultsTo(os.Stdout, result, format)
}
```

The text renderer prints a section for errors and another for warnings — each only when non-empty. A final line summarizes pass/fail and the warning count.

```bash
sed -n '44,72p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/output.go
```

```output
// printTextResults displays human-readable validation results.
func printTextResults(writer io.Writer, result ValidationResult) {
	if len(result.Errors) > 0 {
		fmt.Fprintln(writer, "\n❌ Errors:")

		for _, err := range result.Errors {
			fmt.Fprintf(writer, "  • %s\n", err)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Fprintln(writer, "\n⚠️  Warnings:")

		for _, warning := range result.Warnings {
			fmt.Fprintf(writer, "  • %s\n", warning)
		}
	}

	if result.IsValid() {
		fmt.Fprintln(writer, "\n✅ Manifest validation passed!")

		if len(result.Warnings) > 0 {
			fmt.Fprintf(writer, "   (but has %d warning(s) to consider)\n", len(result.Warnings))
		}
	} else {
		fmt.Fprintf(writer, "\n❌ Validation failed with %d error(s) and %d warning(s)\n",
			len(result.Errors), len(result.Warnings))
	}
}
```

The JSON renderer marshals an `OutputResult` with two-space indentation. Marshal failures (highly unlikely with this shape) fall back to a hand-written JSON literal so the writer always gets some valid JSON.

```bash
sed -n '74,91p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/output.go
```

```output
// printJSONResults displays JSON-formatted validation results.
func printJSONResults(writer io.Writer, result ValidationResult) {
	output := OutputResult{
		Valid:    result.IsValid(),
		Errors:   result.Errors,
		Warnings: result.Warnings,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(writer, `{"valid":false,"errors":["Failed to marshal JSON output: %v"]}`, err)

		return
	}

	_, _ = writer.Write(jsonData)
	_, _ = writer.Write([]byte("\n"))
}
```

## Tests

Tests follow a table-driven pattern with `t.Parallel()` at every level. Each test case is built by a focused factory function and listed in `getValidationTestCases()`.

```bash
sed -n '15,23p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/validator_test.go
```

```output
// getValidationTestCases returns test cases for manifest validation.
func getValidationTestCases() []TestCase {
	return []TestCase{
		getValidManifestTestCase(),
		getInvalidManifestTestCase(),
		getMissingRequiredFieldsTestCase(),
		getWarningOnlyTestCase(),
	}
}
```

The "invalid manifest" case is the most informative — it intentionally trips nearly every rule at once and asserts the exact error/warning counts (9 errors, 2 warnings).

```bash
sed -n '44,61p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/validator_test.go
```

```output
func getInvalidManifestTestCase() TestCase {
	return TestCase{
		name: "invalid manifest",
		manifest: Manifest{
			ID:            "bad-manifest-obsidian-plugin",
			Name:          "Bad Obsidian Plugin",
			Description:   "This plugin for Obsidian has a long description exceeding limits",
			Author:        "author@example.com",
			Version:       "1.0.0-beta",
			MinAppVersion: "1.0.0-final",
			IsDesktopOnly: false,
			AuthorURL:     "https://obsidian.md",
			FundingURL:    "https://obsidian.md/pricing",
		},
		wantErrors:   9,
		wantWarnings: 2,
	}
}
```

The test runner iterates the table and runs each as a subtest in parallel. A `checkValidationResult` helper does the count assertions and logs each individual message on mismatch — making test failures self-diagnosing.

```bash
sed -n '101,134p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/validator_test.go
```

```output
func TestValidateManifest(t *testing.T) {
	t.Parallel()

	testCases := getValidationTestCases()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateManifest(&testCase.manifest)
			checkValidationResult(t, result, testCase.wantErrors, testCase.wantWarnings)
		})
	}
}

// checkValidationResult verifies validation result matches expected values.
func checkValidationResult(t *testing.T, result ValidationResult, wantErrors, wantWarnings int) {
	t.Helper()

	if len(result.Errors) != wantErrors {
		t.Errorf("ValidateManifest() errors = %d, want %d", len(result.Errors), wantErrors)

		for i, err := range result.Errors {
			t.Logf("Error %d: %s", i+1, err)
		}
	}

	if len(result.Warnings) != wantWarnings {
		t.Errorf("ValidateManifest() warnings = %d, want %d", len(result.Warnings), wantWarnings)

		for i, warn := range result.Warnings {
			t.Logf("Warning %d: %s", i+1, warn)
		}
	}
}
```

Output tests in `output_test.go` write to a `bytes.Buffer` via `PrintResultsTo`. The JSON test round-trips the buffer back through `json.Unmarshal` and asserts on the typed struct — a clean way to verify shape without string-matching JSON.

```bash
sed -n '10,41p' /Users/markayers/source/philoserf/obsidian-tools/validate-plugin-manifest/output_test.go
```

```output
func TestJSONOutput(t *testing.T) {
	t.Parallel()

	result := ValidationResult{
		Errors:   []string{},
		Warnings: []string{},
	}
	result.AddError("Test error")
	result.AddWarning("Test warning")

	buffer := &bytes.Buffer{}
	PrintResultsTo(buffer, result, JSONOutput)

	// Parse the JSON output
	var output OutputResult
	if err := json.Unmarshal(buffer.Bytes(), &output); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify the parsed JSON
	if output.Valid != false {
		t.Errorf("Expected valid=false, got %v", output.Valid)
	}

	if len(output.Errors) != 1 || output.Errors[0] != "Test error" {
		t.Errorf("Expected errors=[Test error], got %v", output.Errors)
	}

	if len(output.Warnings) != 1 || output.Warnings[0] != "Test warning" {
		t.Errorf("Expected warnings=[Test warning], got %v", output.Warnings)
	}
}
```

## Build & run

Task targets wrap the standard Go commands. The binary lands in the repo root (the CLI is then invoked with `--manifest <path>`).

```bash
sed -n '16,30p' /Users/markayers/source/philoserf/obsidian-tools/Taskfile.yml
```

```output
  build:
    desc: Build the validator binary
    cmds:
      - go build -o {{.BINARY_NAME}} {{.BUILD_DIR}}
    sources:
      - "{{.BUILD_DIR}}/**/*.go"
      - "go.mod"
    generates:
      - "{{.BINARY_NAME}}"

  test:
    desc: Run tests (extra flags via `--`, e.g. `task test -- -cover`)
    cmds:
      - go test {{.CLI_ARGS}} ./...

```

## Adding a new rule

The recipe is short:

1. Write `validateThing(manifest *Manifest, result *ValidationResult)`.
2. Inside, call `result.AddError(...)`, `result.AddErrorf(...)`, `result.AddWarning(...)`, or `result.AddWarningf(...)` as needed.
3. Add a call to `validateThing(manifest, &result)` in `ValidateManifest`.
4. Extend a test case (or add a new one) to cover both the passing and failing path.

The split between fatal errors and informational warnings is a deliberate API contract — `IsValid()` watches errors only, and `main()` keys its exit code off that.

## Adding a new tool

Per the project layout, a new tool is a sibling directory at the repo root with its own `package main` and `main()` entry. The current build target in `Taskfile.yml` is wired to a single binary; multi-tool builds would require extending `build:` (or adding a `build:all` aggregator as the project guide suggests).

