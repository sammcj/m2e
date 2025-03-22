import React, { useState, useEffect, useRef } from 'react';
import './HighlightedTextarea.css';

/**
 * A completely rebuilt textarea component that highlights American words and smart quotes
 * using a contenteditable div for perfect alignment
 */
function HighlightedTextarea({
    value,
    onChange,
    placeholder,
    dictionary,
    normaliseSmartQuotes,
    smartQuotesMap
}) {
    const [highlightedText, setHighlightedText] = useState('');
    const [showPlaceholder, setShowPlaceholder] = useState(!value);
    const contentEditableRef = useRef(null);

    // Escape HTML special characters
    const escapeHtml = (text) => {
        if (!text) return '';
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    };

    // Handle input changes in the contenteditable div
    const handleInput = (e) => {
        const text = e.target.innerText;
        setShowPlaceholder(!text);

        // Call the onChange handler with the new text
        if (onChange) {
            onChange({ target: { value: text } });
        }
    };

    // Handle paste events to strip formatting and maintain cursor position
    const handlePaste = (e) => {
        e.preventDefault();

        // Get the plain text from the clipboard
        const text = e.clipboardData.getData('text/plain');

        // Insert the text at the current cursor position
        // Using the modern way instead of the deprecated execCommand
        const selection = window.getSelection();
        if (selection && selection.rangeCount > 0) {
            try {
                // Delete any selected text
                selection.deleteFromDocument();

                // Insert the new text
                const range = selection.getRangeAt(0);
                const textNode = document.createTextNode(text);
                range.insertNode(textNode);

                // Move the cursor to the end of the inserted text
                range.setStartAfter(textNode);
                range.setEndAfter(textNode);
                selection.removeAllRanges();
                selection.addRange(range);

                // Trigger the input event to update the value
                const inputEvent = new Event('input', { bubbles: true });
                e.target.dispatchEvent(inputEvent);
            } catch (error) {
                // Fallback to the old way if there's an error
                document.execCommand('insertText', false, text);
            }
        } else {
            // Fallback to the old way if there's no selection
            document.execCommand('insertText', false, text);
        }
    };

    // Handle keydown events
    const handleKeyDown = (e) => {
        // Handle tab key
        if (e.key === 'Tab') {
            e.preventDefault();

            // Insert tab at the current cursor position
            const selection = window.getSelection();
            if (selection && selection.rangeCount > 0) {
                try {
                    // Delete any selected text
                    selection.deleteFromDocument();

                    // Insert the tab character
                    const range = selection.getRangeAt(0);
                    const textNode = document.createTextNode('\t');
                    range.insertNode(textNode);

                    // Move the cursor to the end of the inserted text
                    range.setStartAfter(textNode);
                    range.setEndAfter(textNode);
                    selection.removeAllRanges();
                    selection.addRange(range);

                    // Trigger the input event to update the value
                    const inputEvent = new Event('input', { bubbles: true });
                    e.target.dispatchEvent(inputEvent);
                } catch (error) {
                    // Fallback to the old way if there's an error
                    document.execCommand('insertText', false, '\t');
                }
            } else {
                // Fallback to the old way if there's no selection
                document.execCommand('insertText', false, '\t');
            }
        }
    };

    // Update highlighting whenever relevant props change
    useEffect(() => {
        if (!value) {
            setHighlightedText('');
            setShowPlaceholder(true);
            return;
        }

        setShowPlaceholder(false);

        // Collect all items to highlight
        const highlightItems = [];

        // Add American words to highlight
        if (dictionary && Object.keys(dictionary).length > 0) {
            // Get all American words (values in the dictionary)
            const americanWords = [];
            for (const britishWord in dictionary) {
                const americanWord = dictionary[britishWord];
                if (americanWord) {
                    americanWords.push(americanWord);
                }
            }

            // Find all occurrences of American words in the text
            if (americanWords.length > 0) {
                for (const americanWord of americanWords) {
                    // Create a regex to match this American word with word boundaries
                    const wordRegex = new RegExp(`\\b${americanWord}\\b`, 'gi');

                    // Find all matches in the original text
                    let match;
                    while ((match = wordRegex.exec(value)) !== null) {
                        highlightItems.push({
                            index: match.index,
                            length: match[0].length,
                            text: match[0],
                            type: 'word'
                        });
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
    }, [value, dictionary, normaliseSmartQuotes, smartQuotesMap]);

    // We don't need to sync the contenteditable div with the value prop
    // because we're using dangerouslySetInnerHTML to set the content

    // We'll skip the click handler since it's causing TypeScript errors
    // The contenteditable div should automatically get focus when clicked

    return (
        <div className="highlighted-textarea-container">
            <div
                ref={contentEditableRef}
                className="editable-content"
                contentEditable
                onInput={handleInput}
                onPaste={handlePaste}
                onKeyDown={handleKeyDown}
                dangerouslySetInnerHTML={{ __html: highlightedText }}
            />
            {showPlaceholder && placeholder && (
                <div className="placeholder">{placeholder}</div>
            )}
        </div>
    );
}

export default HighlightedTextarea;
