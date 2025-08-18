// Package report provides functionality for generating reports and formatted output
// from text conversion operations
package report

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

// OutputFormat represents the different output formats available
type OutputFormat string

const (
	OutputDiff     OutputFormat = "diff"
	OutputText     OutputFormat = "text"
	OutputMarkdown OutputFormat = "markdown"
	OutputStats    OutputFormat = "stats"
)

// ReportOptions configures the report generation
type ReportOptions struct {
	ShowDiff     bool
	ShowText     bool
	ShowMarkdown bool
	ShowStats    bool
	ExitOnChange bool
	Width        int
}

// DefaultOptions returns sensible default options for report generation
func DefaultOptions() ReportOptions {
	return ReportOptions{
		ShowDiff:     false,
		ShowText:     false,
		ShowMarkdown: true,
		ShowStats:    true,
		ExitOnChange: false,
		Width:        80,
	}
}

// ChangeStats represents statistics about changes made during conversion
type ChangeStats struct {
	TotalWords      int
	SpellingChanges int
	UnitConversions int
	QuoteChanges    int
	ChangedWords    []WordChange
	ChangedUnits    []UnitChange
}

// WordChange represents a single spelling change
type WordChange struct {
	Original string
	Changed  string
	Position int
}

// UnitChange represents a single unit conversion
type UnitChange struct {
	Original string
	Changed  string
	Position int
	UnitType string
}

// FileResult represents the result of processing a single file
type FileResult struct {
	FilePath   string
	Original   string
	Converted  string
	Stats      ChangeStats
	HasChanges bool
	Error      error
}

// Reporter handles the generation and formatting of conversion reports
type Reporter struct {
	options   ReportOptions
	glamour   *glamour.TermRenderer
	hasChange bool
}

// NewReporter creates a new Reporter with the given options
func NewReporter(options ReportOptions) (*Reporter, error) {
	// Create glamour renderer with auto-style detection and word wrapping
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(options.Width),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	return &Reporter{
		options: options,
		glamour: renderer,
	}, nil
}

// GenerateReport creates a comprehensive report from the conversion results
func (r *Reporter) GenerateReport(original, converted string, stats ChangeStats) (string, error) {
	var output strings.Builder
	r.hasChange = r.calculateHasChange(stats)

	if r.options.ShowDiff {
		diff, err := r.generateDiff(original, converted)
		if err != nil {
			return "", fmt.Errorf("failed to generate diff: %w", err)
		}
		output.WriteString(diff)
		output.WriteString("\n\n")
	}

	if r.options.ShowText {
		output.WriteString("## Converted Text\n\n")
		output.WriteString("```\n")
		output.WriteString(converted)
		output.WriteString("\n```\n\n")
	}

	if r.options.ShowMarkdown {
		markdown, err := r.generateMarkdownOutput(converted)
		if err != nil {
			return "", fmt.Errorf("failed to generate markdown output: %w", err)
		}
		output.WriteString(markdown)
		output.WriteString("\n\n")
	}

	if r.options.ShowStats {
		statsOutput := r.generateStatsOutput(stats)
		output.WriteString(statsOutput)
	}

	return output.String(), nil
}

// GenerateMultiFileReport creates a comprehensive report from multiple file conversion results
func (r *Reporter) GenerateMultiFileReport(results []FileResult) (string, error) {
	var output strings.Builder

	// Calculate overall statistics
	totalFiles := len(results)
	changedFiles := 0
	var allStats ChangeStats
	var errorFiles []string

	for _, result := range results {
		if result.Error != nil {
			errorFiles = append(errorFiles, result.FilePath)
			continue
		}

		if result.HasChanges {
			changedFiles++
		}

		// Aggregate stats
		allStats.TotalWords += result.Stats.TotalWords
		allStats.SpellingChanges += result.Stats.SpellingChanges
		allStats.UnitConversions += result.Stats.UnitConversions
		allStats.QuoteChanges += result.Stats.QuoteChanges
		allStats.ChangedWords = append(allStats.ChangedWords, result.Stats.ChangedWords...)
		allStats.ChangedUnits = append(allStats.ChangedUnits, result.Stats.ChangedUnits...)
	}

	r.hasChange = changedFiles > 0

	// Generate summary first
	output.WriteString("# Multi-File Conversion Report\n\n")
	output.WriteString(fmt.Sprintf("📁 **Files processed:** %d\n", totalFiles))
	output.WriteString(fmt.Sprintf("📝 **Files with changes:** %d\n", changedFiles))

	if len(errorFiles) > 0 {
		output.WriteString(fmt.Sprintf("⚠️  **Files with errors:** %d\n", len(errorFiles)))
	}

	if !r.hasChange {
		output.WriteString("\n✅ **No changes required** - all files are already in international English format.\n\n")
	} else {
		output.WriteString("\n")
		output.WriteString(r.generateAggregateStats(allStats))
		output.WriteString("\n")
	}

	// Show error files if any
	if len(errorFiles) > 0 {
		output.WriteString("## ⚠️ Errors\n\n")
		for _, filePath := range errorFiles {
			output.WriteString(fmt.Sprintf("- `%s`\n", filePath))
		}
		output.WriteString("\n")
	}

	// Generate individual file reports for files with changes
	if r.hasChange {
		output.WriteString("## 📄 Individual File Reports\n\n")

		for _, result := range results {
			if result.Error != nil || !result.HasChanges {
				continue
			}

			fileReport, err := r.generateSingleFileReport(result)
			if err != nil {
				return "", fmt.Errorf("failed to generate report for file %s: %w", result.FilePath, err)
			}

			output.WriteString(fileReport)
			output.WriteString("\n---\n\n")
		}
	}

	return output.String(), nil
}

// generateSingleFileReport creates a report for a single file
func (r *Reporter) generateSingleFileReport(result FileResult) (string, error) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("### 📄 %s\n\n", result.FilePath))

	if r.options.ShowDiff {
		diff, err := r.generateDiff(result.Original, result.Converted)
		if err != nil {
			return "", fmt.Errorf("failed to generate diff: %w", err)
		}
		output.WriteString(diff)
		output.WriteString("\n")
	}

	if r.options.ShowText {
		output.WriteString("**Converted text:**\n\n")
		output.WriteString("```\n")
		output.WriteString(result.Converted)
		output.WriteString("\n```\n\n")
	}

	if r.options.ShowMarkdown {
		markdown, err := r.generateMarkdownOutput(result.Converted)
		if err != nil {
			return "", fmt.Errorf("failed to generate markdown output: %w", err)
		}
		output.WriteString("**Converted content:**\n\n")
		output.WriteString(markdown)
		output.WriteString("\n")
	}

	if r.options.ShowStats {
		statsOutput := r.generateFileStats(result.Stats)
		output.WriteString(statsOutput)
	}

	return output.String(), nil
}

// generateAggregateStats creates aggregate statistics for multiple files
func (r *Reporter) generateAggregateStats(stats ChangeStats) string {
	var output strings.Builder

	output.WriteString("## 📊 Aggregate Statistics\n\n")
	output.WriteString(fmt.Sprintf("📊 **Total words processed:** %d\n", stats.TotalWords))

	if stats.SpellingChanges > 0 {
		output.WriteString(fmt.Sprintf("🔤 **Total spelling changes:** %d\n", stats.SpellingChanges))
	}

	if stats.UnitConversions > 0 {
		output.WriteString(fmt.Sprintf("📏 **Total unit conversions:** %d\n", stats.UnitConversions))
	}

	if stats.QuoteChanges > 0 {
		output.WriteString(fmt.Sprintf("❝ **Total quote normalizations:** %d\n", stats.QuoteChanges))
	}

	return output.String()
}

// generateFileStats creates statistics for a single file
func (r *Reporter) generateFileStats(stats ChangeStats) string {
	var output strings.Builder

	output.WriteString("**File statistics:**\n\n")
	output.WriteString(fmt.Sprintf("- Words processed: %d\n", stats.TotalWords))

	if stats.SpellingChanges > 0 {
		output.WriteString(fmt.Sprintf("- Spelling changes: %d\n", stats.SpellingChanges))
	}

	if stats.UnitConversions > 0 {
		output.WriteString(fmt.Sprintf("- Unit conversions: %d\n", stats.UnitConversions))
	}

	if stats.QuoteChanges > 0 {
		output.WriteString(fmt.Sprintf("- Quote normalizations: %d\n", stats.QuoteChanges))
	}

	// Show some example changes
	if len(stats.ChangedWords) > 0 {
		output.WriteString("\n**Spelling changes:**\n")
		maxShow := 5
		for i, change := range stats.ChangedWords {
			if i >= maxShow {
				remaining := len(stats.ChangedWords) - maxShow
				output.WriteString(fmt.Sprintf("- ... and %d more\n", remaining))
				break
			}
			output.WriteString(fmt.Sprintf("- `%s` → `%s`\n", change.Original, change.Changed))
		}
	}

	if len(stats.ChangedUnits) > 0 {
		output.WriteString("\n**Unit conversions:**\n")
		maxShow := 5
		for i, change := range stats.ChangedUnits {
			if i >= maxShow {
				remaining := len(stats.ChangedUnits) - maxShow
				output.WriteString(fmt.Sprintf("- ... and %d more\n", remaining))
				break
			}
			output.WriteString(fmt.Sprintf("- `%s` → `%s` (%s)\n",
				change.Original, change.Changed, change.UnitType))
		}
	}

	return output.String()
}

// ShouldExitWithError returns true if the tool should exit with code 1
func (r *Reporter) ShouldExitWithError() bool {
	return r.options.ExitOnChange && r.hasChange
}

// HasChanges returns true if any changes were detected
func (r *Reporter) HasChanges() bool {
	return r.hasChange
}

// calculateHasChange determines if any changes were made
func (r *Reporter) calculateHasChange(stats ChangeStats) bool {
	return stats.SpellingChanges > 0 || stats.UnitConversions > 0 || stats.QuoteChanges > 0
}

// generateDiff creates a git-style diff output
func (r *Reporter) generateDiff(original, converted string) (string, error) {
	if original == converted {
		return "No changes detected.", nil
	}

	var diff strings.Builder
	diff.WriteString("```diff\n")
	diff.WriteString("--- original\n")
	diff.WriteString("+++ converted\n")

	originalLines := strings.Split(original, "\n")
	convertedLines := strings.Split(converted, "\n")

	// Simple line-by-line diff
	maxLines := len(originalLines)
	if len(convertedLines) > maxLines {
		maxLines = len(convertedLines)
	}

	for i := 0; i < maxLines; i++ {
		var origLine, convLine string
		if i < len(originalLines) {
			origLine = originalLines[i]
		}
		if i < len(convertedLines) {
			convLine = convertedLines[i]
		}

		if origLine != convLine {
			if origLine != "" {
				diff.WriteString("- ")
				diff.WriteString(origLine)
				diff.WriteString("\n")
			}
			if convLine != "" {
				diff.WriteString("+ ")
				diff.WriteString(convLine)
				diff.WriteString("\n")
			}
		}
	}

	diff.WriteString("```\n")
	return diff.String(), nil
}

// generateMarkdownOutput renders the converted text as formatted markdown
func (r *Reporter) generateMarkdownOutput(converted string) (string, error) {
	// Check if the content is already markdown by looking for markdown syntax
	isMarkdown := strings.Contains(converted, "#") ||
		strings.Contains(converted, "*") ||
		strings.Contains(converted, "[") ||
		strings.Contains(converted, "`")

	var content string
	if isMarkdown {
		content = converted
	} else {
		// Wrap plain text in a code block for better formatting
		content = fmt.Sprintf("```\n%s\n```", converted)
	}

	rendered, err := r.glamour.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return rendered, nil
}

// generateStatsOutput creates a formatted statistics summary
func (r *Reporter) generateStatsOutput(stats ChangeStats) string {
	var output strings.Builder

	output.WriteString("## Conversion Summary\n\n")

	if !r.hasChange {
		output.WriteString("✅ No changes required - text is already in international English format.\n")
		return output.String()
	}

	output.WriteString(fmt.Sprintf("📊 **Total words processed:** %d\n", stats.TotalWords))

	if stats.SpellingChanges > 0 {
		output.WriteString(fmt.Sprintf("🔤 **Spelling changes needed:** %d\n", stats.SpellingChanges))
	}

	if stats.UnitConversions > 0 {
		output.WriteString(fmt.Sprintf("📏 **Unit conversions needed:** %d\n", stats.UnitConversions))
	}

	if stats.QuoteChanges > 0 {
		output.WriteString(fmt.Sprintf("❝ **Quote normalizations needed:** %d\n", stats.QuoteChanges))
	}

	// Show detailed changes if there are any
	if len(stats.ChangedWords) > 0 {
		output.WriteString("\n### Spelling Changes\n")
		for _, change := range stats.ChangedWords {
			output.WriteString(fmt.Sprintf("- `%s` → `%s`\n", change.Original, change.Changed))
		}
	}

	if len(stats.ChangedUnits) > 0 {
		output.WriteString("\n### Unit Conversions\n")
		for _, change := range stats.ChangedUnits {
			output.WriteString(fmt.Sprintf("- `%s` → `%s` (%s)\n",
				change.Original, change.Changed, change.UnitType))
		}
	}

	return output.String()
}
