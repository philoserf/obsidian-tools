package main

import (
	"testing"
)

// TestCase defines a validation test case.
type TestCase struct {
	name         string
	manifest     Manifest
	wantErrors   int
	wantWarnings int
}

// getValidationTestCases returns test cases for manifest validation.
func getValidationTestCases() []TestCase {
	return []TestCase{
		getValidManifestTestCase(),
		getInvalidManifestTestCase(),
		getMissingRequiredFieldsTestCase(),
		getWarningOnlyTestCase(),
	}
}

func getValidManifestTestCase() TestCase {
	return TestCase{
		name: "valid manifest",
		manifest: Manifest{
			ID:            "good-manifest",
			Name:          "Good Manifest",
			Description:   "A good manifest",
			Author:        "The Author",
			Version:       "1.0.0",
			MinAppVersion: "1.0.0",
			IsDesktopOnly: false,
			AuthorURL:     "https://example.com/",
			FundingURL:    "https://example.com/funding",
		},
		wantErrors:   0,
		wantWarnings: 0,
	}
}

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

func getMissingRequiredFieldsTestCase() TestCase {
	return TestCase{
		name: "missing required fields",
		manifest: Manifest{
			ID:            "",
			Name:          "",
			Version:       "",
			Description:   "",
			Author:        "",
			MinAppVersion: "",
			IsDesktopOnly: false,
			AuthorURL:     "",
			FundingURL:    "",
		},
		wantErrors:   6, // ID, Name, Description, MinAppVersion, Author, Version
		wantWarnings: 0,
	}
}

func getWarningOnlyTestCase() TestCase {
	return TestCase{
		name: "warning only",
		manifest: Manifest{
			ID:            "good-id",
			Name:          "Good Name",
			Description:   "This plugin has a description with this plugin phrase",
			Author:        "contact@example.com",
			Version:       "1.0.0",
			MinAppVersion: "1.0.0",
			IsDesktopOnly: false,
			AuthorURL:     "https://example.com",
			FundingURL:    "",
		},
		wantErrors:   0,
		wantWarnings: 2, // Email in author and "this plugin" phrase
	}
}

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

func TestValidationResult(t *testing.T) {
	t.Parallel()
	t.Run("empty result is valid", func(t *testing.T) {
		t.Parallel()

		result := ValidationResult{
			Errors:   []string{},
			Warnings: []string{},
		}
		if !result.IsValid() {
			t.Errorf("Empty ValidationResult should be valid")
		}
	})

	t.Run("result with errors is invalid", func(t *testing.T) {
		t.Parallel()

		result := ValidationResult{
			Errors:   []string{},
			Warnings: []string{},
		}
		result.AddError("Test error")

		if result.IsValid() {
			t.Errorf("ValidationResult with errors should be invalid")
		}
	})

	t.Run("result with warnings only is valid", func(t *testing.T) {
		t.Parallel()

		result := ValidationResult{
			Errors:   []string{},
			Warnings: []string{},
		}
		result.AddWarning("Test warning")

		if !result.IsValid() {
			t.Errorf("ValidationResult with only warnings should be valid")
		}
	})
}
