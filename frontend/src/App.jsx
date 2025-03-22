import React, { useState, useEffect, useRef } from 'react';
import './App.css';
import { ConvertToBritish, ConvertToAmerican, HandleDroppedFile, SaveConvertedFile, GetCurrentFilePath, ClearCurrentFile } from "../wailsjs/go/main/App";

function App() {
    const [freedomText, setAmericanText] = useState('');
    const [britishText, setBritishText] = useState('');
    const [normaliseSmartQuotes, setNormaliseSmartQuotes] = useState(true);
    const [currentFilePath, setCurrentFilePath] = useState('');
    const [dragActive, setDragActive] = useState(false);
    const [fileError, setFileError] = useState('');

    const appContainerRef = useRef(null);

    // Check if a file was opened with the app
    useEffect(() => {
        GetCurrentFilePath().then(path => {
            if (path) {
                setCurrentFilePath(path);
            }
        });
    }, []);

    // Update the American English text area
    const updateAmericanText = (e) => {
        setAmericanText(e.target.value);
    };

    // Update the British English text area
    const updateBritishText = (e) => {
        setBritishText(e.target.value);
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
                    ConvertToBritish(content, normaliseSmartQuotes).then(result => {
                        setBritishText(result);
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

        ConvertToBritish(freedomText, normaliseSmartQuotes).then((result) => {
            setBritishText(result);
        });
    };

    // Convert from British to American English
    const handleBritishToAmerican = () => {
        if (!britishText.trim()) return;

        ConvertToAmerican(britishText, normaliseSmartQuotes).then((result) => {
            setAmericanText(result);
        });
    };

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
                console.error('Failed to copy text: ', err);
                alert('Failed to copy text to clipboard');
            });
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
            <header className="app-header">
                <div className="header-content">
                    <h3>Murican English Conversion</h3>
                    <div className="header-controls">
                        <label className="checkbox-container">
                            <input
                                type="checkbox"
                                checked={normaliseSmartQuotes}
                                onChange={toggleNormaliseSmartQuotes}
                            />
                            <span className="checkbox-text">Normalise smart quotes and dashes</span>
                        </label>

                        {currentFilePath && (
                            <div className="file-controls">
                                <span className="current-file">Current file: {currentFilePath}</span>
                                <button className="save-file-button" onClick={handleSaveFile}>
                                    Save File
                                </button>
                                <button className="clear-file-button" onClick={handleClearFile}>
                                    Clear File
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </header>

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

            <div className="controls-row">
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
                </div>

                <div className="language-label">'Murican</div>
                <button
                    className="clear-button"
                    onClick={handleClear}
                >
                    Clear All
                </button>

                <div className="language-label">English</div>
                <div className="button-group">
                    <button
                        className="convert-button british-to-american"
                        onClick={handleBritishToAmerican}
                    >
                        Convert to 'Murican
                    </button>
                    <button
                        className="copy-button"
                        onClick={() => copyToClipboard(britishText)}
                    >
                        Copy
                    </button>
                </div>
            </div>

            <div className="converter-container">
                <div className="text-column">
                    <textarea
                        className="text-area"
                        value={freedomText}
                        onChange={updateAmericanText}
                        placeholder="Enter freedom text here or drop a text file..."
                    />
                </div>

                <div className="text-column">
                    <textarea
                        className="text-area"
                        value={britishText}
                        onChange={updateBritishText}
                        placeholder="Enter British English text here..."
                    />
                </div>
            </div>
        </div>
    );
}

export default App;
