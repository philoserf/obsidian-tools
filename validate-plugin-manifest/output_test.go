package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

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

func TestTextOutput(t *testing.T) {
	t.Parallel()

	t.Run("with errors and warnings", func(t *testing.T) {
		t.Parallel()

		result := ValidationResult{
			Errors:   []string{},
			Warnings: []string{},
		}
		result.AddError("Test error")
		result.AddWarning("Test warning")

		buffer := &bytes.Buffer{}
		PrintResultsTo(buffer, result, TextOutput)

		output := buffer.String()

		if !strings.Contains(output, "❌ Errors:") {
			t.Errorf("Expected output to contain '❌ Errors:', got: %s", output)
		}

		if !strings.Contains(output, "Test error") {
			t.Errorf("Expected output to contain 'Test error', got: %s", output)
		}

		if !strings.Contains(output, "⚠️  Warnings:") {
			t.Errorf("Expected output to contain '⚠️  Warnings:', got: %s", output)
		}

		if !strings.Contains(output, "Test warning") {
			t.Errorf("Expected output to contain 'Test warning', got: %s", output)
		}
	})

	t.Run("with success", func(t *testing.T) {
		t.Parallel()

		result := ValidationResult{
			Errors:   []string{},
			Warnings: []string{},
		} // Empty result = valid

		buffer := &bytes.Buffer{}
		PrintResultsTo(buffer, result, TextOutput)

		output := buffer.String()

		if !strings.Contains(output, "✅ Manifest validation passed!") {
			t.Errorf("Expected output to contain '✅ Manifest validation passed!', got: %s", output)
		}
	})
}
