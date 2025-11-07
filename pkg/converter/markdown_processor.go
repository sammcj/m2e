package converter

import (
	"fmt"
	"regexp"
	"strings"
)

// MarkdownProcessor handles preservation of markdown formatting during conversion
type MarkdownProcessor struct {
	boldAsteriskPattern     *regexp.Regexp
	boldUnderscorePattern   *regexp.Regexp
	italicAsteriskPattern   *regexp.Regexp
	italicUnderscorePattern *regexp.Regexp
	linkPattern             *regexp.Regexp
}

// NewMarkdownProcessor creates a new markdown processor
func NewMarkdownProcessor() *MarkdownProcessor {
	return &MarkdownProcessor{
		boldAsteriskPattern:     regexp.MustCompile(`\*\*([^*]+)\*\*`),
		boldUnderscorePattern:   regexp.MustCompile(`__([^_]+)__`),
		italicAsteriskPattern:   regexp.MustCompile(`(\s|^)\*([^\s*][^*]*?)\*(\s|$|[,.!?;:])`),
		italicUnderscorePattern: regexp.MustCompile(`(\s|^)_([^\s_][^_]*?)_(\s|$|[,.!?;:])`),
		linkPattern:             regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`),
	}
}

// ProcessWithMarkdown converts text while preserving markdown formatting
func (mp *MarkdownProcessor) ProcessWithMarkdown(text string, convertFunc func(string) string) string {
	if text == "" {
		return text
	}

	// Check if text contains any markdown patterns
	hasMarkdown := mp.hasMarkdownPatterns(text)

	// If no markdown, just convert the text directly
	if !hasMarkdown {
		return convertFunc(text)
	}

	result := text

	// Step 1: Extract bold/italic formatting first (so it works inside links too)
	type formattingInfo struct {
		placeholder string
		text        string
		prefix      string
		suffix      string
	}
	var formatting []formattingInfo
	fmtIdx := 0

	// Handle ** bold
	result = mp.boldAsteriskPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := mp.boldAsteriskPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		placeholder := fmt.Sprintf("XMDBLDX%dXMDBLDX", fmtIdx)
		formatting = append(formatting, formattingInfo{placeholder, parts[1], "**", "**"})
		fmtIdx++
		return placeholder
	})

	// Handle __ bold
	result = mp.boldUnderscorePattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := mp.boldUnderscorePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		placeholder := fmt.Sprintf("XMDBLDX%dXMDBLDX", fmtIdx)
		formatting = append(formatting, formattingInfo{placeholder, parts[1], "__", "__"})
		fmtIdx++
		return placeholder
	})

	// Handle * italic
	result = mp.italicAsteriskPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := mp.italicAsteriskPattern.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		placeholder := fmt.Sprintf("XMDITLX%dXMDITLX", fmtIdx)
		formatting = append(formatting, formattingInfo{placeholder, parts[2], parts[1] + "*", "*" + parts[3]})
		fmtIdx++
		return parts[1] + placeholder + parts[3]
	})

	// Handle _ italic
	result = mp.italicUnderscorePattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := mp.italicUnderscorePattern.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		placeholder := fmt.Sprintf("XMDITLX%dXMDITLX", fmtIdx)
		formatting = append(formatting, formattingInfo{placeholder, parts[2], parts[1] + "_", "_" + parts[3]})
		fmtIdx++
		return parts[1] + placeholder + parts[3]
	})

	// Step 2: Extract markdown links (which may now contain formatting placeholders)
	type linkInfo struct {
		placeholder string
		linkText    string
		url         string
	}
	var links []linkInfo
	linkIdx := 0
	result = mp.linkPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := mp.linkPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		placeholder := fmt.Sprintf("XMDLINKX%dXMDLINKX", linkIdx)
		links = append(links, linkInfo{placeholder, parts[1], parts[2]})
		linkIdx++
		return placeholder
	})

	// Step 3: Convert remaining text
	result = convertFunc(result)

	// Step 4: Restore formatting with converted text
	// Store converted text to avoid redundant conversions
	convertedFormatting := make(map[string]string)
	for _, fmt := range formatting {
		convertedText := convertFunc(fmt.text)
		convertedFormatting[fmt.placeholder] = convertedText

		// For bold, use full prefix+text+suffix
		// For italic, just use the marker without the whitespace (already in place)
		var restored string
		if fmt.prefix == "**" || fmt.prefix == "__" {
			restored = fmt.prefix + convertedText + fmt.suffix
		} else {
			// Italic - extract just the marker from prefix/suffix
			marker := "*"
			if strings.Contains(fmt.prefix, "_") {
				marker = "_"
			}
			restored = marker + convertedText + marker
		}
		result = strings.ReplaceAll(result, fmt.placeholder, restored)
	}

	// Step 5: Restore links - link text may contain formatting placeholders, so restore those too
	for _, link := range links {
		// The link text might have formatting placeholders - restore them first
		linkText := link.linkText
		for _, fmt := range formatting {
			if strings.Contains(linkText, fmt.placeholder) {
				// Reuse already converted text from map
				convertedText := convertedFormatting[fmt.placeholder]

				// Use same logic as step 4 to handle bold vs italic consistently
				var restored string
				if fmt.prefix == "**" || fmt.prefix == "__" {
					restored = fmt.prefix + convertedText + fmt.suffix
				} else {
					// Italic - extract just the marker from prefix/suffix
					marker := "*"
					if strings.Contains(fmt.prefix, "_") {
						marker = "_"
					}
					restored = marker + convertedText + marker
				}
				linkText = strings.ReplaceAll(linkText, fmt.placeholder, restored)
			}
		}

		// Convert any remaining plain text in the link
		convertedLinkText := convertFunc(linkText)
		markdownLink := "[" + convertedLinkText + "](" + link.url + ")"
		result = strings.ReplaceAll(result, link.placeholder, markdownLink)
	}

	return result
}

// hasMarkdownPatterns checks if text contains any markdown formatting
func (mp *MarkdownProcessor) hasMarkdownPatterns(text string) bool {
	// Check for markdown links
	if strings.Contains(text, "](") {
		return true
	}

	// Check for bold markers
	if strings.Contains(text, "**") || strings.Contains(text, "__") {
		return true
	}

	// Check for potential italic markers (more careful check)
	// Count asterisks and underscores
	asteriskCount := strings.Count(text, "*")
	underscoreCount := strings.Count(text, "_")

	// If we have pairs of asterisks or underscores, might be italic
	if asteriskCount >= 2 || underscoreCount >= 2 {
		return true
	}

	return false
}
