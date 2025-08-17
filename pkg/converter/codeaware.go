package converter

import (
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/lexers"
)

// CodeBlock represents a detected code block in text
type CodeBlock struct {
	Start    int    // Start position in original text
	End      int    // End position in original text
	Language string // Detected language (if any)
	Content  string // Raw content of the code block
	IsCode   bool   // true if this is code, false if it's regular text
}

// TextSegment represents a segment of text that can be either code or regular text
type TextSegment struct {
	Content  string
	IsCode   bool
	Language string
}

// DetectCodeBlocks detects and extracts code blocks from mixed text
func (c *Converter) DetectCodeBlocks(text string) []CodeBlock {
	var blocks []CodeBlock

	// First, detect markdown fenced code blocks
	blocks = append(blocks, c.detectMarkdownCodeBlocks(text)...)

	// Then detect inline code blocks
	blocks = append(blocks, c.detectInlineCode(text)...)

	// Finally, try to detect raw code if no markdown blocks were found
	if len(blocks) == 0 {
		blocks = append(blocks, c.detectRawCode(text)...)
	}

	// Fill in text segments between code blocks
	blocks = c.fillTextSegments(text, blocks)

	return blocks
}

// detectMarkdownCodeBlocks finds fenced code blocks (``` and ~~~)
func (c *Converter) detectMarkdownCodeBlocks(text string) []CodeBlock {
	var blocks []CodeBlock

	// Regex for fenced code blocks with optional language specification
	// We'll handle backtick and tilde fences separately since Go doesn't support backreferences
	backtickRegex := regexp.MustCompile(`(?ms)^` + "`{3}" + `([a-zA-Z0-9+-]*)\n?(.*?)\n?` + "`{3}" + `\s*$`)
	tildeRegex := regexp.MustCompile(`(?ms)^~~~([a-zA-Z0-9+-]*)\n?(.*?)\n?~~~\s*$`)

	// Process backtick fenced blocks
	matches := backtickRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		start := match[0]
		end := match[1]
		langStart := match[2]
		langEnd := match[3]
		contentStart := match[4]
		contentEnd := match[5]

		language := ""
		if langStart >= 0 && langEnd > langStart {
			language = text[langStart:langEnd]
		}

		content := ""
		if contentStart >= 0 && contentEnd > contentStart {
			content = text[contentStart:contentEnd]
		}

		blocks = append(blocks, CodeBlock{
			Start:    start,
			End:      end,
			Language: language,
			Content:  content,
			IsCode:   true,
		})
	}

	// Process tilde fenced blocks
	matches = tildeRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		start := match[0]
		end := match[1]
		langStart := match[2]
		langEnd := match[3]
		contentStart := match[4]
		contentEnd := match[5]

		language := ""
		if langStart >= 0 && langEnd > langStart {
			language = text[langStart:langEnd]
		}

		content := ""
		if contentStart >= 0 && contentEnd > contentStart {
			content = text[contentStart:contentEnd]
		}

		blocks = append(blocks, CodeBlock{
			Start:    start,
			End:      end,
			Language: language,
			Content:  content,
			IsCode:   true,
		})
	}

	return blocks
}

// detectInlineCode finds inline code blocks (`code`)
func (c *Converter) detectInlineCode(text string) []CodeBlock {
	var blocks []CodeBlock

	// Regex for inline code (single backticks)
	inlineRegex := regexp.MustCompile("`([^`\n]+)`")

	matches := inlineRegex.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		start := match[0]
		end := match[1]
		contentStart := match[2]
		contentEnd := match[3]

		content := text[contentStart:contentEnd]

		blocks = append(blocks, CodeBlock{
			Start:    start,
			End:      end,
			Language: "", // Unknown for inline code
			Content:  content,
			IsCode:   true,
		})
	}

	return blocks
}

// detectRawCode attempts to detect if the entire text is code
func (c *Converter) detectRawCode(text string) []CodeBlock {
	// Try to detect the language using Chroma
	lexer := lexers.Analyse(text)
	if lexer != nil {
		config := lexer.Config()
		if config != nil && config.Name != "plaintext" && config.Name != "Text" {
			// Looks like code, treat the entire text as a code block
			return []CodeBlock{
				{
					Start:    0,
					End:      len(text),
					Language: strings.ToLower(config.Name),
					Content:  text,
					IsCode:   true,
				},
			}
		}
	}

	return nil
}

// fillTextSegments fills in text segments between code blocks
func (c *Converter) fillTextSegments(text string, codeBlocks []CodeBlock) []CodeBlock {
	if len(codeBlocks) == 0 {
		// No code blocks, entire text is regular text
		return []CodeBlock{
			{
				Start:    0,
				End:      len(text),
				Language: "",
				Content:  text,
				IsCode:   false,
			},
		}
	}

	var allBlocks []CodeBlock
	lastEnd := 0

	for _, block := range codeBlocks {
		// Add text segment before this code block
		if block.Start > lastEnd {
			textContent := text[lastEnd:block.Start]
			if strings.TrimSpace(textContent) != "" {
				allBlocks = append(allBlocks, CodeBlock{
					Start:    lastEnd,
					End:      block.Start,
					Language: "",
					Content:  textContent,
					IsCode:   false,
				})
			}
		}

		// Add the code block
		allBlocks = append(allBlocks, block)
		lastEnd = block.End
	}

	// Add remaining text after the last code block
	if lastEnd < len(text) {
		textContent := text[lastEnd:]
		if strings.TrimSpace(textContent) != "" {
			allBlocks = append(allBlocks, CodeBlock{
				Start:    lastEnd,
				End:      len(text),
				Language: "",
				Content:  textContent,
				IsCode:   false,
			})
		}
	}

	return allBlocks
}

// ExtractComments extracts comment text from code using Chroma
func (c *Converter) ExtractComments(code, language string) []CommentBlock {
	// For now, use manual extraction as it handles newlines better
	// TODO: Fix Chroma extraction to include proper boundaries
	return c.extractCommentsManually(code)
}

// CommentBlock represents a comment within code
type CommentBlock struct {
	Start   int    // Start position in code
	End     int    // End position in code
	Content string // Comment text
}

// extractCommentsManually provides fallback comment detection using regex
func (c *Converter) extractCommentsManually(code string) []CommentBlock {
	return c.extractCommentsManuallyWithConversion(code, false, false)
}

// extractCommentsManuallyWithConversion provides comment detection with optional unit conversion
func (c *Converter) extractCommentsManuallyWithConversion(code string, convertUnits bool, normaliseSmartQuotes bool) []CommentBlock {
	var comments []CommentBlock

	// Line comment patterns that should include newlines
	lineCommentPatterns := []*regexp.Regexp{
		regexp.MustCompile(`//.*?(?:\n|$)`), // Line comments: // comment with newline
		regexp.MustCompile(`#.*?(?:\n|$)`),  // Hash comments: # comment with newline
	}

	// Block comment patterns (already include their boundaries)
	blockCommentPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?s)/\*.*?\*/`), // Block comments: /* comment */
		regexp.MustCompile(`(?s)""".*?"""`), // Python docstrings: """comment"""
		regexp.MustCompile(`(?s)'''.*?'''`), // Python docstrings: '''comment'''
		regexp.MustCompile(`<!--.*?-->`),    // HTML comments: <!-- comment -->
	}

	// Find line comments (include newline if present)
	for _, pattern := range lineCommentPatterns {
		matches := pattern.FindAllStringIndex(code, -1)
		for _, match := range matches {
			start := match[0]
			end := match[1]
			content := code[start:end]

			// Remove trailing newline from content for processing, but keep the position
			content = strings.TrimSuffix(content, "\n")

			// Apply unit conversion if requested
			if convertUnits && c.unitProcessor != nil && c.unitProcessor.IsEnabled() {
				// Apply spelling conversion first
				convertedContent := c.ConvertToBritishSimple(content, normaliseSmartQuotes)
				// Then apply unit conversion
				convertedContent = c.unitProcessor.ProcessText(convertedContent, false, "")
				content = convertedContent
			}

			comments = append(comments, CommentBlock{
				Start:   start,
				End:     end,
				Content: content,
			})
		}
	}

	// Find block comments
	for _, pattern := range blockCommentPatterns {
		matches := pattern.FindAllStringIndex(code, -1)
		for _, match := range matches {
			start := match[0]
			end := match[1]
			content := code[start:end]

			// Apply unit conversion if requested
			if convertUnits && c.unitProcessor != nil && c.unitProcessor.IsEnabled() {
				// Apply spelling conversion first
				convertedContent := c.ConvertToBritishSimple(content, normaliseSmartQuotes)
				// Then apply unit conversion
				convertedContent = c.unitProcessor.ProcessText(convertedContent, false, "")
				content = convertedContent
			}

			comments = append(comments, CommentBlock{
				Start:   start,
				End:     end,
				Content: content,
			})
		}
	}

	return comments
}

// ProcessCodeAware processes text with code-awareness
func (c *Converter) ProcessCodeAware(text string, normaliseSmartQuotes bool) string {
	// Simple approach: check if we have any code blocks at all
	// If not, use regular conversion with both spelling and unit conversion
	if !c.containsCodeBlocks(text) {
		// Apply spelling conversion first
		result := c.ConvertToBritishSimple(text, normaliseSmartQuotes)
		// Then apply unit conversion
		if c.unitProcessor != nil && c.unitProcessor.IsEnabled() {
			result = c.unitProcessor.ProcessText(result, false, "")
		}
		return result
	}

	// Process the text by converting only non-code parts
	return c.processTextWithCodeBlocks(text, normaliseSmartQuotes)
}

// containsCodeBlocks does a quick check for code blocks
func (c *Converter) containsCodeBlocks(text string) bool {
	// Check for fenced code blocks
	if strings.Contains(text, "```") || strings.Contains(text, "~~~") {
		return true
	}
	// Check for inline code
	if strings.Contains(text, "`") {
		return true
	}
	return false
}

// processTextWithCodeBlocks processes text while preserving code blocks
func (c *Converter) processTextWithCodeBlocks(text string, normaliseSmartQuotes bool) string {
	// Handle fenced code blocks first (``` and ~~~)
	result := c.processFencedCodeBlocks(text, normaliseSmartQuotes)
	// Then handle inline code
	result = c.processInlineCode(result, normaliseSmartQuotes)
	return result
}

// processFencedCodeBlocks handles markdown fenced code blocks
func (c *Converter) processFencedCodeBlocks(text string, normaliseSmartQuotes bool) string {
	// Split by fenced code blocks and process each part
	parts := c.splitByFencedBlocks(text)
	var result strings.Builder

	for _, part := range parts {
		if part.IsCode {
			// This is a code block - only convert comments
			convertedContent := c.convertCommentsInCode(part.Content, part.Language, normaliseSmartQuotes)

			// Reconstruct the full block with original fence type
			if part.Language != "" {
				result.WriteString(part.FenceType + part.Language + "\n" + convertedContent + "\n" + part.FenceType)
			} else {
				result.WriteString(part.FenceType + "\n" + convertedContent + "\n" + part.FenceType)
			}
		} else {
			// Regular text - apply both spelling and unit conversion
			converted := c.ConvertToBritishSimple(part.Content, normaliseSmartQuotes)
			// Then apply unit conversion
			if c.unitProcessor != nil && c.unitProcessor.IsEnabled() {
				converted = c.unitProcessor.ProcessText(converted, false, "")
			}
			result.WriteString(converted)
		}
	}

	return result.String()
}

// processInlineCode handles inline code blocks
func (c *Converter) processInlineCode(text string, normaliseSmartQuotes bool) string {
	// If the text contains fenced code blocks, don't process inline code
	// because the fenced blocks have already been processed correctly
	if strings.Contains(text, "```") || strings.Contains(text, "~~~") {
		return text
	}

	// Use regex to find and preserve inline code while converting surrounding text
	inlineRegex := regexp.MustCompile("`([^`\n]+)`")

	// Check if there are any inline code matches
	if !inlineRegex.MatchString(text) {
		// No inline code, process as regular text
		converted := c.ConvertToBritishSimple(text, normaliseSmartQuotes)
		if c.unitProcessor != nil && c.unitProcessor.IsEnabled() {
			converted = c.unitProcessor.ProcessText(converted, false, "")
		}
		return converted
	}

	// Split the text by inline code blocks and process the non-code parts
	parts := inlineRegex.Split(text, -1)
	matches := inlineRegex.FindAllString(text, -1)

	var result strings.Builder
	for i, part := range parts {
		if part != "" {
			// This is regular text - apply both spelling and unit conversion
			converted := c.ConvertToBritishSimple(part, normaliseSmartQuotes)
			if c.unitProcessor != nil && c.unitProcessor.IsEnabled() {
				converted = c.unitProcessor.ProcessText(converted, false, "")
			}
			result.WriteString(converted)
		}

		// Add back the inline code block if it exists
		if i < len(matches) {
			// For inline code, preserve it as-is (don't convert anything)
			// Inline code should not be processed for spelling changes
			result.WriteString(matches[i])
		}
	}

	return result.String()
}

// TextPart represents a part of text that can be code or regular text
type TextPart struct {
	Content   string
	IsCode    bool
	Language  string
	FenceType string // "```" or "~~~" for fenced code blocks
}

// splitByFencedBlocks splits text by fenced code blocks
func (c *Converter) splitByFencedBlocks(text string) []TextPart {
	var parts []TextPart

	// Simple regex to match fenced blocks
	fenceRegex := regexp.MustCompile(`(?s)` + "`{3}" + `([a-zA-Z0-9+-]*)\n?(.*?)\n?` + "`{3}" + `|(?s)~~~([a-zA-Z0-9+-]*)\n?(.*?)\n?~~~`)

	lastEnd := 0
	matches := fenceRegex.FindAllStringSubmatchIndex(text, -1)

	for _, match := range matches {
		start := match[0]
		end := match[1]

		// Add text before this code block
		if start > lastEnd {
			textContent := text[lastEnd:start]
			if textContent != "" {
				parts = append(parts, TextPart{
					Content: textContent,
					IsCode:  false,
				})
			}
		}

		// Determine language, content, and fence type
		var language, content, fenceType string
		if match[2] >= 0 { // Backtick fence
			language = text[match[2]:match[3]]
			content = text[match[4]:match[5]]
			fenceType = "```"
		} else if match[6] >= 0 { // Tilde fence
			language = text[match[6]:match[7]]
			content = text[match[8]:match[9]]
			fenceType = "~~~"
		}

		// Add the code block
		parts = append(parts, TextPart{
			Content:   content,
			IsCode:    true,
			Language:  language,
			FenceType: fenceType,
		})

		lastEnd = end
	}

	// Add remaining text
	if lastEnd < len(text) {
		remaining := text[lastEnd:]
		if remaining != "" {
			parts = append(parts, TextPart{
				Content: remaining,
				IsCode:  false,
			})
		}
	}

	// If no code blocks found, return the entire text as one part
	if len(parts) == 0 {
		parts = append(parts, TextPart{
			Content: text,
			IsCode:  false,
		})
	}

	return parts
}

// convertCommentsInCode converts only comments within code
func (c *Converter) convertCommentsInCode(code, language string, normaliseSmartQuotes bool) string {
	comments := c.ExtractComments(code, language)

	if len(comments) == 0 {
		return code
	}

	// Use a simple replacement approach: replace each comment one by one
	// working backwards so positions don't shift
	for i := len(comments) - 1; i >= 0; i-- {
		comment := comments[i]

		// Get the original comment block (including any trailing newline)
		originalBlock := code[comment.Start:comment.End]

		// Convert just the comment content (without newline) - apply both spelling and unit conversion
		converted := c.ConvertToBritishSimple(comment.Content, normaliseSmartQuotes)
		// Then apply unit conversion
		if c.unitProcessor != nil && c.unitProcessor.IsEnabled() {
			converted = c.unitProcessor.ProcessText(converted, false, "")
		}

		// If the original block had a trailing newline, preserve it
		if strings.HasSuffix(originalBlock, "\n") {
			converted += "\n"
		}

		// Replace this comment in the code
		before := code[:comment.Start]
		after := code[comment.End:]
		code = before + converted + after
	}

	return code
}
