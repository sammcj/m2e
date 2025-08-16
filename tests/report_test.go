package tests

import (
	"strings"
	"testing"

	"github.com/sammcj/m2e/pkg/report"
)

func TestNewReporter(t *testing.T) {
	options := report.DefaultOptions()
	reporter, err := report.NewReporter(options)

	if err != nil {
		t.Fatalf("Expected no error creating reporter, got: %v", err)
	}

	if reporter == nil {
		t.Fatal("Expected reporter to be created, got nil")
	}
}

func TestDefaultOptions(t *testing.T) {
	options := report.DefaultOptions()

	if options.ShowDiff {
		t.Error("Expected ShowDiff to be false by default")
	}

	if options.ShowText {
		t.Error("Expected ShowText to be false by default")
	}

	if !options.ShowMarkdown {
		t.Error("Expected ShowMarkdown to be true by default")
	}

	if !options.ShowStats {
		t.Error("Expected ShowStats to be true by default")
	}

	if options.ExitOnChange {
		t.Error("Expected ExitOnChange to be false by default")
	}

	if options.Width != 80 {
		t.Errorf("Expected Width to be 80 by default, got %d", options.Width)
	}
}

func TestGenerateReport_NoChanges(t *testing.T) {
	options := report.ReportOptions{
		ShowMarkdown: true,
		ShowStats:    true,
		Width:        80,
	}

	reporter, err := report.NewReporter(options)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	original := "This is already correct text."
	converted := original
	stats := report.ChangeStats{
		TotalWords:      5,
		SpellingChanges: 0,
		UnitConversions: 0,
		QuoteChanges:    0,
	}

	output, err := reporter.GenerateReport(original, converted, stats)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	if !strings.Contains(output, "No changes required") {
		t.Error("Expected 'No changes required' message in output")
	}

	if reporter.HasChanges() {
		t.Error("Expected HasChanges to be false when no changes detected")
	}

	if reporter.ShouldExitWithError() {
		t.Error("Expected ShouldExitWithError to be false when no changes detected")
	}
}

func TestGenerateReport_WithChanges(t *testing.T) {
	options := report.ReportOptions{
		ShowStats:    true,
		ShowMarkdown: true,
		ExitOnChange: true,
		Width:        80,
	}

	reporter, err := report.NewReporter(options)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	original := "I love color and humor."
	converted := "I love colour and humour."
	stats := report.ChangeStats{
		TotalWords:      5,
		SpellingChanges: 2,
		UnitConversions: 0,
		QuoteChanges:    0,
		ChangedWords: []report.WordChange{
			{Original: "color", Changed: "colour", Position: 7},
			{Original: "humor", Changed: "humour", Position: 17},
		},
	}

	output, err := reporter.GenerateReport(original, converted, stats)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	if !strings.Contains(output, "Spelling changes:** 2") {
		t.Error("Expected spelling changes count in output")
	}

	if !strings.Contains(output, "color` → `colour") {
		t.Error("Expected specific word change in output")
	}

	if !reporter.HasChanges() {
		t.Error("Expected HasChanges to be true when changes detected")
	}

	if !reporter.ShouldExitWithError() {
		t.Error("Expected ShouldExitWithError to be true when ExitOnChange is true and changes detected")
	}
}

func TestGenerateReport_WithDiff(t *testing.T) {
	options := report.ReportOptions{
		ShowDiff: true,
		Width:    80,
	}

	reporter, err := report.NewReporter(options)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	original := "I love color."
	converted := "I love colour."
	stats := report.ChangeStats{
		TotalWords:      3,
		SpellingChanges: 1,
	}

	output, err := reporter.GenerateReport(original, converted, stats)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	if !strings.Contains(output, "```diff") {
		t.Error("Expected diff block in output")
	}

	if !strings.Contains(output, "- I love color.") {
		t.Error("Expected original line with minus prefix in diff")
	}

	if !strings.Contains(output, "+ I love colour.") {
		t.Error("Expected converted line with plus prefix in diff")
	}
}

func TestGenerateReport_WithText(t *testing.T) {
	options := report.ReportOptions{
		ShowText: true,
		Width:    80,
	}

	reporter, err := report.NewReporter(options)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	original := "I love color."
	converted := "I love colour."
	stats := report.ChangeStats{}

	output, err := reporter.GenerateReport(original, converted, stats)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	if !strings.Contains(output, "## Converted Text") {
		t.Error("Expected 'Converted Text' header in output")
	}

	if !strings.Contains(output, "I love colour.") {
		t.Error("Expected converted text in output")
	}
}

func TestAnalyser_CountWords(t *testing.T) {
	analyser := report.NewAnalyser(map[string]string{
		"color": "colour",
		"humor": "humour",
	})

	text := "This is a test with five words."
	stats := analyser.AnalyseChanges(text, text)

	if stats.TotalWords != 7 {
		t.Errorf("Expected 7 words, got %d", stats.TotalWords)
	}
}

func TestAnalyser_SpellingChanges(t *testing.T) {
	americanWords := map[string]string{
		"color": "colour",
		"humor": "humour",
	}

	analyser := report.NewAnalyser(americanWords)

	original := "I love color and humor."
	converted := "I love colour and humour."
	stats := analyser.AnalyseChanges(original, converted)

	if stats.SpellingChanges != 2 {
		t.Errorf("Expected 2 spelling changes, got %d", stats.SpellingChanges)
	}

	if len(stats.ChangedWords) != 2 {
		t.Errorf("Expected 2 changed words, got %d", len(stats.ChangedWords))
	}

	// Check first change
	if stats.ChangedWords[0].Original != "color" || stats.ChangedWords[0].Changed != "colour" {
		t.Errorf("Unexpected first word change: %s -> %s",
			stats.ChangedWords[0].Original, stats.ChangedWords[0].Changed)
	}
}

func TestAnalyser_QuoteChanges(t *testing.T) {
	analyser := report.NewAnalyser(map[string]string{})

	original := "\u201cSmart quotes\u201d and \u2018apostrophes\u2019 with em\u2014dashes."
	converted := "\"Smart quotes\" and 'apostrophes' with em-dashes."
	stats := analyser.AnalyseChanges(original, converted)

	if stats.QuoteChanges < 3 { // Should detect smart quotes and em-dash
		t.Errorf("Expected at least 3 quote changes, got %d", stats.QuoteChanges)
	}
}

func TestUnitTypeDetection(t *testing.T) {
	analyser := report.NewAnalyser(map[string]string{})

	testCases := []struct {
		unit     string
		expected string
	}{
		{"12 feet", "length"},
		{"5 pounds", "mass"},
		{"2 gallons", "volume"},
		{"75°F", "temperature"},
		{"100 square feet", "area"},
		{"unknown unit", "unknown"},
	}

	for _, tc := range testCases {
		// This is testing the internal determineUnitType method indirectly
		// through the unit conversion analysis
		original := "The measurement is " + tc.unit + "."
		converted := original // No actual conversion for this test
		stats := analyser.AnalyseChanges(original, converted)

		// Since we're not doing actual conversions, we mainly test that
		// the analyser doesn't crash on different unit types
		if stats.TotalWords == 0 {
			t.Errorf("Analyser failed to process text with unit: %s", tc.unit)
		}
	}
}
