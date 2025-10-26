// Package main provides a validator for Obsidian plugin manifests.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	version = "0.1.0"
)

var (
	// Pre-compiled regular expressions for better performance.
	idRegex      = regexp.MustCompile(`^[a-z0-9-_]+$`)
	emailRegex   = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	versionRegex = regexp.MustCompile(`^[0-9.]+$`)
)

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
func (vr *ValidationResult) AddErrorf(format string, args ...interface{}) {
	vr.Errors = append(vr.Errors, fmt.Sprintf(format, args...))
}

// AddWarning adds a warning message to the validation result.
func (vr *ValidationResult) AddWarning(message string) {
	vr.Warnings = append(vr.Warnings, message)
}

// AddWarningf adds a formatted warning message to the validation result.
func (vr *ValidationResult) AddWarningf(format string, args ...interface{}) {
	vr.Warnings = append(vr.Warnings, fmt.Sprintf(format, args...))
}

// IsValid returns true if there are no validation errors.
func (vr *ValidationResult) IsValid() bool {
	return len(vr.Errors) == 0
}

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

// validateURLs checks author and funding URL rules.
func validateURLs(manifest *Manifest, result *ValidationResult) {
	if manifest.AuthorURL == "https://obsidian.md" {
		result.AddError("Author URL should not point to Obsidian website")
	}

	if manifest.FundingURL != "" && manifest.FundingURL == "https://obsidian.md/pricing" {
		result.AddError("Funding URL should not point to Obsidian pricing")
	}
}

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

// PrintResults moved to output.go

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
