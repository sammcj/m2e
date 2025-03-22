import React, { useState, useEffect, useMemo } from 'react';
import './HighlightedTextarea.css';

/**
 * A textarea component that highlights words that match a dictionary
 */
function HighlightedTextarea({
    value,
    onChange,
    placeholder,
    dictionary,
    normaliseSmartQuotes,
    smartQuotesMap
}) {
    // Sync scroll positions between the textarea and the highlight div
    const handleScroll = (e) => {
        const textarea = e.target;
        const highlightLayer = textarea.previousSibling;
        highlightLayer.scrollTop = textarea.scrollTop;
        highlightLayer.scrollLeft = textarea.scrollLeft;
    };

    // Function to escape HTML special characters
    const escapeHtml = (text) => {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    };

    // Memoize the highlighted HTML to prevent unnecessary re-renders
    const highlightedHtml = useMemo(() => {
        if (!value) {
            return '';
        }

        // Skip processing if dictionary is empty
        if (!dictionary || Object.keys(dictionary).length === 0) {
            return escapeHtml(value).replace(/\n/g, '<br>');
        }

        // Create a set of words to highlight (case-insensitive)
        const wordsToHighlight = new Set();
        for (const word in dictionary) {
            wordsToHighlight.add(word.toLowerCase());
        }

        // Create a set of smart quotes to highlight if normaliseSmartQuotes is enabled
        const quotesToHighlight = new Set();
        if (normaliseSmartQuotes && smartQuotesMap) {
            for (const quote in smartQuotesMap) {
                quotesToHighlight.add(quote);
            }
        }

        // Split the text into words and preserve whitespace
        const words = value.split(/(\s+)/);

        // Highlight words that match the dictionary
        let html = '';
        for (const word of words) {
            const trimmedWord = word.trim().toLowerCase();
            const punctuationMatch = trimmedWord.match(/^(\w+)([^\w]*)$/);

            let wordToCheck = trimmedWord;
            let punctuation = '';

            if (punctuationMatch) {
                wordToCheck = punctuationMatch[1];
                punctuation = punctuationMatch[2];
            }

            // Check if the word should be highlighted
            if (wordsToHighlight.has(wordToCheck)) {
                html += `<span class="highlight-word">${escapeHtml(word)}</span>`;
            }
            // Check if the word contains smart quotes to highlight
            else if (normaliseSmartQuotes) {
                let hasQuote = false;
                let highlightedWord = escapeHtml(word);

                for (const quote of quotesToHighlight) {
                    if (word.includes(quote)) {
                        hasQuote = true;
                        highlightedWord = highlightedWord.replace(
                            new RegExp(quote, 'g'),
                            `<span class="highlight-quote">${quote}</span>`
                        );
                    }
                }

                html += hasQuote ? highlightedWord : highlightedWord;
            }
            else {
                html += escapeHtml(word);
            }
        }

        // Replace newlines with <br> for proper display
        return html.replace(/\n/g, '<br>');
    }, [value, dictionary, normaliseSmartQuotes, smartQuotesMap]);

    return (
        <div className="highlighted-textarea-container">
            <div
                className="highlight-layer"
                dangerouslySetInnerHTML={{ __html: highlightedHtml || placeholder }}
            />
            <textarea
                className="text-input"
                value={value}
                onChange={onChange}
                placeholder={placeholder}
                onScroll={handleScroll}
                style={{ caretColor: 'black' }} // Make the cursor visible
            />
        </div>
    );
}

export default HighlightedTextarea;
