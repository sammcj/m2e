import React, { useState, useEffect, useRef, useCallback } from 'react';
import './App.css';
import { ConvertToBritish, ConvertToBritishWithUnits, HandleDroppedFile, SaveConvertedFile, GetCurrentFilePath, ClearCurrentFile, GetUnitProcessingStatus, SetUnitProcessingEnabled } from "../wailsjs/go/main/App";
import HighlightedTextarea from './components/HighlightedTextarea';

function App() {
    const [freedomText, setAmericanText] = useState('');
    const [britishText, setBritishText] = useState('');
    const [normaliseSmartQuotes, setNormaliseSmartQuotes] = useState(true);
    const [syntaxHighlighting, setSyntaxHighlighting] = useState(false);
    const [convertUnits, setConvertUnits] = useState(false);
    const [currentFilePath, setCurrentFilePath] = useState('');
    const [dragActive, setDragActive] = useState(false);
    const [fileError, setFileError] = useState('');
    const [americanToBritishDict, setAmericanToBritishDict] = useState({});
    const [smartQuotesMap, setSmartQuotesMap] = useState({});
    const [isTranslating, setIsTranslating] = useState(false); // Flag to prevent infinite loops
    const translationTimerRef = useRef(null); // For JS, this is fine; for TS, use: useRef<number | null>(null)
    const [showEagle, setShowEagle] = useState(false); // State to control eagle animation

    const appContainerRef = useRef(null);
    // Create a ref for the eagle element
    const eagleRef = useRef(null);

    // Check if a file was opened with the app and load dictionaries
    useEffect(() => {
        // Check for file path
        GetCurrentFilePath().then(path => {
            if (path) {
                setCurrentFilePath(path);
            }
        });

        // Get unit processing status
        GetUnitProcessingStatus().then(status => {
            setConvertUnits(status);
        }).catch(err => {
            console.error('Error getting unit processing status:', err);
        });

        // Get the dictionary directly from the backend
        // We've added methods to the backend to expose the dictionary
        import("../wailsjs/go/main/App").then(({ GetAmericanToBritishDictionary }) => {
            // Get the American to British dictionary
            GetAmericanToBritishDictionary().then(dict => {
                // American to British dictionary loaded successfully
                setAmericanToBritishDict(dict);
            }).catch(err => {
                // Handle error loading American to British dictionary
            });
        }).catch(err => {
            // Handle error importing App methods
        });

        const smartQuotesMap = {
            "\u201C": "\"", // Left double quote
            "\u201D": "\"", // Right double quote
            "\u2018": "'",  // Left single quote
            "\u2019": "'",  // Right single quote
            "\u2013": "-",  // En-dash
            "\u2014": "-"  // Em-dash
        };

        setSmartQuotesMap(smartQuotesMap);
    }, []);

    // Update the American English text area and automatically translate
    const updateAmericanText = (e) => {
        const newText = e.target.value;
        setAmericanText(newText);

        // Clear existing timer
        if (translationTimerRef.current) {
            clearTimeout(translationTimerRef.current);
        }

        // Set new debounced timer - only translate after user stops typing
        const timer = setTimeout(() => {
            if (newText.trim()) {
                setIsTranslating(true);
                ConvertToBritishWithUnits(newText, normaliseSmartQuotes, convertUnits).then((result) => {
                    setBritishText(result);
                    setIsTranslating(false);
                });
            } else {
                setBritishText('');
            }
        }, 500); // 500ms delay

        translationTimerRef.current = timer;
    };

    // Update the British English text area
    const updateBritishText = (e) => {
        const newText = e.target.value;
        setBritishText(newText);
    };

    // Handle drag events
    const handleDrag = (e) => {
        e.preventDefault();
        e.stopPropagation();

        if (e.type === 'dragenter' || e.type === 'dragover') {
            setDragActive(true);
        } else if (e.type === 'dragleave') {
            setDragActive(false);
        }
    };

    // Handle drop event
    const handleDrop = (e) => {
        e.preventDefault();
        e.stopPropagation();
        setDragActive(false);
        setFileError('');

        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            const file = e.dataTransfer.files[0];

            // Read the file directly using FileReader
            const reader = new FileReader();
            reader.onload = (event) => {
                if (event.target && event.target.result) {
                    const content = event.target.result.toString();
                    setAmericanText(content);

                    // Store the file path if available
                    if (file.path) {
                        setCurrentFilePath(file.path);
                    }

                    // Automatically convert to British English
                    setIsTranslating(true);
                    ConvertToBritishWithUnits(content, normaliseSmartQuotes, convertUnits).then(result => {
                        setBritishText(result);
                        setIsTranslating(false);
                    });
                }
            };

            reader.onerror = () => {
                setFileError("Error reading file. Please try again.");
            };

            reader.readAsText(file);
        }
    };

    // Save the converted file
    const handleSaveFile = () => {
        if (currentFilePath && britishText) {
            SaveConvertedFile(britishText).then(() => {
                setCurrentFilePath('');
                alert('File saved successfully!');
            }).catch(err => {
                setFileError(`Error saving file: ${err.message}`);
            });
        }
    };

    // Clear the current file
    const handleClearFile = () => {
        ClearCurrentFile().then(() => {
            setCurrentFilePath('');
        });
    };

    // Convert from American to British English
    const handleAmericanToBritish = () => {
        if (!freedomText.trim()) return;
        if (isTranslating) return;

        // Trigger eagle animation
        triggerEagleAnimation();

        setIsTranslating(true);
        ConvertToBritishWithUnits(freedomText, normaliseSmartQuotes, convertUnits).then((result) => {
            setBritishText(result);
            setIsTranslating(false);
        });
    };

    // Function to trigger the eagle animation
    const triggerEagleAnimation = useCallback(() => {
        // Hide any existing eagle first
        setShowEagle(false);

        // Then show a new eagle after a brief delay
        setTimeout(() => {
            setShowEagle(true);

            // Hide the eagle after animation completes
            setTimeout(() => {
                setShowEagle(false);
            }, 1500); // Animation duration
        }, 10);
    }, []);


    // Toggle normalise smart quotes option
    const toggleNormaliseSmartQuotes = () => {
        setNormaliseSmartQuotes(!normaliseSmartQuotes);
    };

    // Clear both text areas
    const handleClear = () => {
        setAmericanText('');
        setBritishText('');
    };

    // Copy text to clipboard
    const copyToClipboard = (text) => {
        navigator.clipboard.writeText(text)
            .then(() => {
                alert('Text copied to clipboard!');
            })
            .catch(err => {
                // Handle error copying text to clipboard
                alert('Failed to copy text to clipboard');
            });
    };

    // Reference to the American text area
    const americanTextareaRef = useRef(null);

    // Paste from clipboard - tries multiple approaches
    const pasteFromClipboard = () => {
        try {
            // First, try to focus the textarea
            const textareaElement = document.querySelector('.text-column:first-child .highlighted-textarea-container textarea');
            if (textareaElement) {
                /** @type {HTMLTextAreaElement} */ (textareaElement).focus();

                // Try to execute the paste command
                const successful = document.execCommand('paste');

                if (successful) {
                    // The paste was successful, but we need to manually trigger the onChange event
                    // to update the state and convert the text
                    setTimeout(() => {
                        // Get the text from the textarea
                        const text = /** @type {HTMLTextAreaElement} */ (textareaElement).value;

                        // Update the American text
                        setAmericanText(text);

                        // Automatically convert to British English
                        if (text.trim()) {
                            setIsTranslating(true);
                            ConvertToBritishWithUnits(text, normaliseSmartQuotes, convertUnits).then((result) => {
                                setBritishText(result);
                                setIsTranslating(false);
                            });
                        } else {
                            setBritishText('');
                        }
                    }, 100); // Small delay to ensure the paste has completed

                    return;
                }
            }

            // If execCommand failed or no textarea was found, try the Clipboard API
            navigator.clipboard.readText()
                .then(text => {
                    // Update the American text
                    setAmericanText(text);

                    // Automatically convert to British English
                    if (text.trim()) {
                        setIsTranslating(true);
                        ConvertToBritishWithUnits(text, normaliseSmartQuotes, convertUnits).then((result) => {
                            setBritishText(result);
                            setIsTranslating(false);
                        });
                    } else {
                        setBritishText('');
                    }
                })
                .catch(err => {
                    console.error('Error reading from clipboard:', err);

                    // If all else fails, just focus the textarea
                    if (textareaElement) {
                        /** @type {HTMLTextAreaElement} */ (textareaElement).focus();
                    }
                });
        } catch (err) {
            console.error('Error in paste function:', err);

            // If all else fails, just focus the textarea
            const textareaElement = document.querySelector('.text-column:first-child .highlighted-textarea-container textarea');
            if (textareaElement) {
                /** @type {HTMLTextAreaElement} */ (textareaElement).focus();
            }
        }
    };

    return (
        <div
            className={`app-container ${dragActive ? 'drag-active' : ''}`}
            ref={appContainerRef}
            onDragEnter={handleDrag}
            onDragLeave={handleDrag}
            onDragOver={handleDrag}
            onDrop={handleDrop}
        >
            {currentFilePath && (
                <div className="file-controls-container">
                    <div className="file-controls">
                        <span className="current-file">Current file: {currentFilePath}</span>
                        <button className="save-file-button" onClick={handleSaveFile}>
                            Save File
                        </button>
                        <button className="clear-file-button" onClick={handleClearFile}>
                            Clear File
                        </button>
                    </div>
                </div>
            )}

            {fileError && (
                <div className="error-message">
                    {fileError}
                </div>
            )}

            {dragActive && (
                <div className="drag-overlay">
                    <div className="drag-message">
                        Drop text file here to convert
                    </div>
                </div>
            )}

            {/* Eagle emoji animation */}
            {showEagle && (
                <div
                    ref={eagleRef}
                    className="eagle-emoji eagle-fly"
                >
                    🦅
                </div>
            )}

            <div className="controls-row">
                <h3 className="app-title">Murican English Conversion</h3>
                <div className="button-group">
                    <button
                        className="convert-button american-to-british"
                        onClick={handleAmericanToBritish}
                    >
                        Convert to English
                    </button>
                    <button
                        className="copy-button"
                        onClick={() => copyToClipboard(freedomText)}
                    >
                        Copy
                    </button>
                    <button
                        className="paste-button"
                        onClick={pasteFromClipboard}
                    >
                        Paste
                    </button>
                    <button
                        className="clear-button"
                        onClick={handleClear}
                    >
                        Clear
                    </button>
                </div>
            </div>

            <div className="settings-row">
                <div className="settings-group">
                    <label className="checkbox-label">
                        <input
                            type="checkbox"
                            checked={normaliseSmartQuotes}
                            onChange={toggleNormaliseSmartQuotes}
                        />
                        Normalise Smart Quotes
                    </label>
                    <label className="checkbox-label">
                        <input
                            type="checkbox"
                            checked={syntaxHighlighting}
                            onChange={(e) => setSyntaxHighlighting(e.target.checked)}
                        />
                        Code Syntax Highlighting
                    </label>
                    <label className="checkbox-label">
                        <input
                            type="checkbox"
                            checked={convertUnits}
                            onChange={(e) => {
                                const enabled = e.target.checked;
                                setConvertUnits(enabled);
                                SetUnitProcessingEnabled(enabled);
                            }}
                        />
                        Freedom Unit Conversion
                    </label>
                </div>
            </div>

            <div className="converter-container">
                <div className="text-column">
                    <HighlightedTextarea
                        value={freedomText}
                        onChange={updateAmericanText}
                        onFocus={() => console.log('American textarea focused')}
                        onBlur={() => console.log('American textarea blurred')}
                        placeholder="Enter freedom text here or drop a text file..."
                        dictionary={syntaxHighlighting ? {} : americanToBritishDict}
                        normaliseSmartQuotes={normaliseSmartQuotes}
                        smartQuotesMap={smartQuotesMap}
                        highlightAmericanWords={!syntaxHighlighting} // Only highlight American words if not using syntax highlighting
                        autoFocus={true} // Auto-focus this field when the app launches
                        syntaxHighlighting={syntaxHighlighting}
                        language="auto"
                    />
                </div>

                <div className="text-column">
                    <HighlightedTextarea
                        value={britishText}
                        onChange={updateBritishText}
                        onFocus={() => console.log('British textarea focused')}
                        onBlur={() => console.log('British textarea blurred')}
                        placeholder="English with less Zs will appear here..."
                        dictionary={{}}
                        normaliseSmartQuotes={normaliseSmartQuotes}
                        smartQuotesMap={smartQuotesMap}
                        highlightAmericanWords={!syntaxHighlighting} // Only highlight American words if not using syntax highlighting
                        syntaxHighlighting={syntaxHighlighting}
                        language="auto"
                    />
                </div>
            </div>
        </div>
    );
}

export default App;
