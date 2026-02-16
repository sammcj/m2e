import React, { useState, useEffect, useRef, useMemo } from 'react';
import './HighlightedTextarea.css';
import './SyntaxHighlighting.css';
import { GetSyntaxHighlightedHTML, DetectLanguage } from '../../wailsjs/go/main/App';

/**
 * A completely rebuilt textarea component that highlights American words and smart quotes
 * using a contenteditable div for perfect alignment
 */
function HighlightedTextarea({
    value,
    onChange,
    onFocus,
    onBlur,
    placeholder,
    dictionary,
    normaliseSmartQuotes,
    smartQuotesMap,
    highlightAmericanWords = true, // Default to true for backward compatibility
    autoFocus = false, // Add autoFocus prop with default value
    syntaxHighlighting = false, // Enable syntax highlighting (controlled by parent)
    language = "auto" // Programming language for syntax highlighting
}) {
    const [highlightedText, setHighlightedText] = useState('');

    // Check if the text is inside a markdown code block
    const isInsideMarkdownCodeBlock = (text) => {
        const codeBlockRegex = /```[\s\S]*?```/g;
        return codeBlockRegex.test(text);
    };

    // Escape HTML special characters
    const escapeHtml = (text) => {
        if (!text) return '';
        return text
            .replace(/&/g, '&')
            .replace(/</g, '<')
            .replace(/>/g, '>')
            .replace(/"/g, '"')
            .replace(/'/g, '&#039;');
    };

    // Update highlighting whenever relevant props change
    useEffect(() => {
        if (!value) {
            setHighlightedText('');
            return;
        }

        if (syntaxHighlighting) {
            handleSyntaxHighlighting();
        } else {
            handleWordHighlighting();
        }
    }, [value, dictionary, normaliseSmartQuotes, smartQuotesMap, highlightAmericanWords, syntaxHighlighting, language]);

    // Handle syntax highlighting using Chroma
    const handleSyntaxHighlighting = async () => {
        try {
            let detectedLanguage = language;

            // Auto-detect language if needed
            if (language === "auto") {
                try {
                    detectedLanguage = await DetectLanguage(value);
                } catch (err) {
                    console.warn('Language detection failed:', err);
                    detectedLanguage = "text";
                }
            }

            // If the language is plain text and not in a code block, use word highlighting
            if (detectedLanguage === 'text' && !isInsideMarkdownCodeBlock(value)) {
                handleWordHighlighting();
                return;
            }

            // Get syntax highlighted HTML from backend
            try {
                const syntaxHTML = await GetSyntaxHighlightedHTML(value, detectedLanguage);
                setHighlightedText(syntaxHTML);
                return;
            } catch (err) {
                console.warn('Syntax highlighting failed:', err);
            }
        } catch (err) {
            console.warn('Syntax highlighting error:', err);
        }

        // Fallback to escaped HTML if syntax highlighting fails
        setHighlightedText(escapeHtml(value));
    };

    // Memoise compiled regex patterns keyed on dictionary content
    const compiledWordPatterns = useMemo(() => {
        if (!highlightAmericanWords || !dictionary || Object.keys(dictionary).length === 0) {
            return [];
        }

        const isMuricanSide = Object.keys(dictionary).includes("color");
        const wordsToHighlight = isMuricanSide
            ? Object.keys(dictionary)
            : Object.values(dictionary).filter(Boolean);

        return wordsToHighlight.map(word => {
            const escapedWord = word.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            return [
                new RegExp(`\\b${escapedWord}\\b`, 'gi'),
                new RegExp(`["']${escapedWord}["']`, 'gi'),
                new RegExp(`["']${escapedWord}\\b`, 'gi'),
                new RegExp(`\\b${escapedWord}["']`, 'gi'),
                new RegExp(`\\b${escapedWord}[,.;:!?)]`, 'gi'),
                new RegExp(`[([{]${escapedWord}\\b`, 'gi'),
            ];
        });
    }, [dictionary, highlightAmericanWords]);

    // Handle word and quote highlighting (original logic)
    const handleWordHighlighting = () => {
        // Collect all items to highlight
        const highlightItems = [];

        // Add words to highlight using pre-compiled patterns
        if (compiledWordPatterns.length > 0) {
            for (const patterns of compiledWordPatterns) {
                for (const pattern of patterns) {
                    // Reset lastIndex for global regexes
                    pattern.lastIndex = 0;
                    let match;
                    while ((match = pattern.exec(value)) !== null) {
                        const matchedText = match[0];

                        // Determine the actual word to highlight
                        let startOffset = 0;
                        let endOffset = 0;

                        // Check if the match starts with a non-letter character
                        if (matchedText.length > 0 && !(/[a-zA-Z0-9]/).test(matchedText[0])) {
                            startOffset = 1;
                        }

                        // Check if the match ends with a non-letter character
                        if (matchedText.length > 0 && !(/[a-zA-Z0-9]/).test(matchedText[matchedText.length - 1])) {
                            endOffset = 1;
                        }

                        // Calculate the actual word position and length
                        const actualIndex = match.index + startOffset;
                        const actualLength = matchedText.length - startOffset - endOffset;

                        // Only add if we haven't already added this exact highlight
                        const isDuplicate = highlightItems.some(item =>
                            item.index === actualIndex && item.length === actualLength
                        );

                        if (!isDuplicate) {
                            highlightItems.push({
                                index: actualIndex,
                                length: actualLength,
                                text: matchedText.substring(startOffset, matchedText.length - endOffset),
                                type: 'word'
                            });
                        }
                    }
                }
            }
        }

        // Add smart quotes to highlight
        if (normaliseSmartQuotes && smartQuotesMap && Object.keys(smartQuotesMap).length > 0) {
            // Find all occurrences of smart quotes in the text
            for (const quote of Object.keys(smartQuotesMap)) {
                let index = -1;
                while ((index = value.indexOf(quote, index + 1)) !== -1) {
                    highlightItems.push({
                        index,
                        length: quote.length,
                        text: quote,
                        type: 'quote'
                    });
                }
            }
        }

        // Sort highlight items by index (ascending)
        highlightItems.sort((a, b) => a.index - b.index);

        // Apply highlights
        if (highlightItems.length === 0) {
            // No highlights needed
            setHighlightedText(escapeHtml(value));
            return;
        }

        // Build the highlighted HTML
        let result = '';
        let lastIndex = 0;

        for (const item of highlightItems) {
            // Add text before this highlight
            if (item.index > lastIndex) {
                const beforeText = value.substring(lastIndex, item.index);
                result += escapeHtml(beforeText);
            }

            // Add the highlighted text
            const highlightedText = escapeHtml(item.text);
            const highlightClass = item.type === 'word' ? 'highlight-word' : 'highlight-quote';
            result += `<span class="${highlightClass}">${highlightedText}</span>`;

            // Update the last index
            lastIndex = item.index + item.length;
        }

        // Add any remaining text
        if (lastIndex < value.length) {
            result += escapeHtml(value.substring(lastIndex));
        }

        setHighlightedText(result);
    };

    const textareaRef = useRef(null);
    const backdropRef = useRef(null);

    const handleScroll = () => {
        if (backdropRef.current && textareaRef.current) {
            backdropRef.current.scrollTop = textareaRef.current.scrollTop;
            backdropRef.current.scrollLeft = textareaRef.current.scrollLeft;
        }
    };

    return (
        <div className="highlighted-textarea-container">
            <div ref={backdropRef} className="highlighted-textarea-backdrop">
                <div
                    className="highlights"
                    dangerouslySetInnerHTML={{ __html: highlightedText }}
                />
            </div>
            <textarea
                ref={textareaRef}
                className="highlighted-textarea-input"
                onChange={onChange}
                onScroll={handleScroll}
                value={value}
                placeholder={placeholder}
                onFocus={onFocus}
                onBlur={onBlur}
                autoFocus={autoFocus}
            />
        </div>
    );
}

export default HighlightedTextarea;
