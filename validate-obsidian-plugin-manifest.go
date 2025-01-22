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

type ValidationResult struct {
	Errors   []string
	Warnings []string
}

func (vr *ValidationResult) addError(format string, args ...interface{}) {
	vr.Errors = append(vr.Errors, fmt.Sprintf(format, args...))
}

func (vr *ValidationResult) addWarning(format string, args ...interface{}) {
	vr.Warnings = append(vr.Warnings, fmt.Sprintf(format, args...))
}

func (vr *ValidationResult) isValid() bool {
	return len(vr.Errors) == 0
}

func validateManifest(manifest *Manifest) ValidationResult {
	result := ValidationResult{}

	// Validate ID
	if strings.Contains(strings.ToLower(manifest.ID), "obsidian") {
		result.addError("Plugin ID should not contain the word 'obsidian'")
	}
	if strings.HasSuffix(strings.ToLower(manifest.ID), "plugin") {
		result.addError("Plugin ID should not end with 'plugin'")
	}
	if !regexp.MustCompile(`^[a-z0-9-_]+$`).MatchString(manifest.ID) {
		result.addError("Plugin ID must contain only lowercase alphanumeric characters and dashes")
	}

	// Validate Name
	if strings.Contains(strings.ToLower(manifest.Name), "obsidian") {
		result.addError("Plugin name should not contain the word 'Obsidian'")
	}
	if strings.HasSuffix(strings.ToLower(manifest.Name), "plugin") {
		result.addError("Plugin name should not end with 'Plugin'")
	}

	// Validate Description
	if strings.Contains(strings.ToLower(manifest.Description), "obsidian") {
		result.addError("Description should not contain the word 'Obsidian'")
	}
	if strings.Contains(strings.ToLower(manifest.Description), "this plugin") {
		result.addWarning("Avoid phrases like 'this plugin' in the description")
	}
	if len(manifest.Description) > 250 {
		result.addError("Description should be under 250 characters (currently %d)", len(manifest.Description))
	}

	// Validate Author
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)
	if emailRegex.MatchString(manifest.Author) {
		result.addWarning("Email addresses are discouraged in the author field")
	}

	// Validate Version
	if !regexp.MustCompile(`^[0-9.]+$`).MatchString(manifest.Version) {
		result.addError("Invalid version number format (should only contain numbers and dots)")
	}

	// Validate URLs
	if manifest.AuthorURL == "https://obsidian.md" {
		result.addError("Author URL should not point to Obsidian website")
	}

	if manifest.FundingURL != "" {
		if manifest.FundingURL == "https://obsidian.md/pricing" {
			result.addError("Funding URL should not point to Obsidian pricing")
		}
	}

	// Validate MinAppVersion
	if manifest.MinAppVersion == "" {
		result.addError("MinAppVersion is required")
	} else if !regexp.MustCompile(`^[0-9.]+$`).MatchString(manifest.MinAppVersion) {
		result.addError("Invalid minAppVersion format (should only contain numbers and dots)")
	}

	return result
}

func printResults(result ValidationResult) {
	if len(result.Errors) > 0 {
		fmt.Println("\n❌ Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  • %s\n", err)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\n⚠️  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
	}

	if result.isValid() {
		fmt.Println("\n✅ Manifest validation passed!")
		if len(result.Warnings) > 0 {
			fmt.Printf("   (but has %d warning(s) to consider)\n", len(result.Warnings))
		}
	} else {
		fmt.Printf("\n❌ Validation failed with %d error(s) and %d warning(s)\n",
			len(result.Errors), len(result.Warnings))
	}
}

func main() {
	var manifestPath string
	flag.StringVar(&manifestPath, "manifest", "manifest.json", "Path to manifest.json file")
	flag.Parse()

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
		fmt.Fprintf(os.Stderr, "Error reading manifest.json: %v\n", err)
		os.Exit(1)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing manifest.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📝 Validating manifest for plugin: %s\n", manifest.Name)
	result := validateManifest(&manifest)
	printResults(result)

	if !result.isValid() {
		os.Exit(1)
	}
}
