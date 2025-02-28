package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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
