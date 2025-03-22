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

        // Create a simple HTML version of the text with newlines replaced
        let html = escapeHtml(value).replace(/\n/g, '<br>');

        // Create a set of words to highlight
        const wordsToHighlight = new Set();

        // Add both British and American words to the set
        for (const britishWord in dictionary) {
            wordsToHighlight.add(britishWord.toLowerCase());
            const americanWord = dictionary[britishWord];
            if (americanWord) {
                wordsToHighlight.add(americanWord.toLowerCase());
            }
        }

        // Create a temporary DOM element to work with the HTML
        const tempDiv = document.createElement('div');
        tempDiv.innerHTML = html;

        // Get the text content of the div
        const textContent = tempDiv.textContent || tempDiv.innerText;

        // Find all words in the text
        const wordMatches = textContent.match(/\b[a-zA-Z]+\b/g) || [];

        // For each word, check if it's in the dictionary
        for (const word of wordMatches) {
            if (wordsToHighlight.has(word.toLowerCase()) && word.length > 1) {
                // Replace the word with a highlighted version
                // Use a regex with word boundaries to ensure we only match whole words
                const regex = new RegExp(`\\b${word}\\b`, 'g');
                html = html.replace(regex, `<span class="highlight-word">$&</span>`);
            }
        }

        // Handle smart quotes if enabled
        if (normaliseSmartQuotes && smartQuotesMap) {
            for (const quote in smartQuotesMap) {
                // Escape the quote for use in regex
                const escapedQuote = escapeHtml(quote);

                // Replace all occurrences with highlighted version
                html = html.split(escapedQuote).join(`<span class="highlight-quote">${escapedQuote}</span>`);
            }
        }

        return html;
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
