"use strict";
/**
 * Test fixtures containing various text samples for testing M2E extension
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.unitConversions = exports.expectedDiagnostics = exports.fileContents = exports.britishSpellings = exports.americanSpellings = exports.testTexts = void 0;
exports.testTexts = {
    // Simple cases
    simple: {
        american: 'The color is great',
        british: 'The colour is great',
        changes: 1
    },
    // Multiple spellings
    multiple: {
        american: 'The color of the organization center was analyzed',
        british: 'The colour of the organisation centre was analysed',
        changes: 4
    },
    // Mixed content
    mixed: {
        american: 'The colour of the organization was analyzed', // Already has some British
        british: 'The colour of the organisation was analysed',
        changes: 2
    },
    // No changes needed
    noChanges: {
        american: 'The colour is great and organised',
        british: 'The colour is great and organised',
        changes: 0
    },
    // Code with comments
    jsCode: {
        american: `// The color is great
function analyzeColor(colorValue) {
    // Organize the data
    const center = findCenter(colorValue);
    return center;
}`,
        british: `// The colour is great
function analyzeColor(colorValue) {
    // Organise the data
    const center = findCenter(colorValue);
    return center;
}`,
        changes: 2 // Only comments converted
    },
    // Code comments only
    jsCommentsOnly: {
        american: `// The color is great
function colorAnalyzer() {
    return "color";
}`,
        british: `// The colour is great
function colorAnalyzer() {
    return "color";
}`,
        changes: 1 // Only comment converted, function name and string preserved
    },
    // Python code
    pythonCode: {
        american: `# The color analyzer
def analyze_color(color_value):
    """Analyze the color data."""
    # Organize the data
    return color_value`,
        british: `# The colour analyser
def analyze_color(color_value):
    """Analyse the colour data."""
    # Organise the data
    return color_value`,
        changes: 4 // Comments and docstring converted
    },
    // Markdown document
    markdown: {
        american: `# Color Analysis

This document analyzes the color organization.

## Center of Interest

The center of the color wheel shows organization.`,
        british: `# Colour Analysis

This document analyses the colour organisation.

## Centre of Interest

The centre of the colour wheel shows organisation.`,
        changes: 7
    },
    // Large text sample
    large: {
        american: 'The color organization '.repeat(1000),
        get british() { return this.american.replace(/color/g, 'colour').replace(/organization/g, 'organisation'); },
        changes: 2000
    },
    // Text with units
    withUnits: {
        american: 'The room is 70°F and measures 10 feet by 12 feet',
        british: 'The room is 21°C and measures 3.05m by 3.66m',
        changes: 3 // Temperature + 2 distance measurements
    },
    // Edge cases
    edgeCases: {
        empty: {
            american: '',
            british: '',
            changes: 0
        },
        whitespaceOnly: {
            american: '   \n\t  ',
            british: '   \n\t  ',
            changes: 0
        },
        numbersOnly: {
            american: '123 456 789',
            british: '123 456 789',
            changes: 0
        },
        punctuationOnly: {
            american: '!@#$%^&*()',
            british: '!@#$%^&*()',
            changes: 0
        }
    },
    // Unicode and special characters
    unicode: {
        american: 'The cölor of the organisation is great',
        british: 'The cölour of the organisation is great',
        changes: 1
    },
    // Case variations
    caseVariations: {
        american: 'COLOR Color color COLOR',
        british: 'COLOUR Colour colour COLOUR',
        changes: 4
    },
    // Words that shouldn't be converted (proper nouns, etc.)
    properNouns: {
        american: 'Colorado University Organization (proper noun)',
        british: 'Colorado University Organisation (proper noun)', // Only "Organization" converted
        changes: 1
    }
};
/**
 * American spellings that should be detected in diagnostics
 */
exports.americanSpellings = [
    'color', 'colors', 'colored', 'coloring', 'colorful',
    'organize', 'organizes', 'organized', 'organizing', 'organization',
    'center', 'centers', 'centered', 'centering',
    'analyze', 'analyzes', 'analyzed', 'analyzing', 'analysis',
    'realize', 'realizes', 'realized', 'realizing', 'realization',
    'aluminum', 'honor', 'honors', 'honored', 'honoring',
    'flavor', 'flavors', 'flavored', 'flavoring',
    'neighbor', 'neighbors', 'neighborhood',
    'defense', 'offense', 'license', 'practice'
];
/**
 * British spellings that should NOT be flagged as American
 */
exports.britishSpellings = [
    'colour', 'colours', 'coloured', 'colouring', 'colourful',
    'organise', 'organises', 'organised', 'organising', 'organisation',
    'centre', 'centres', 'centred', 'centring',
    'analyse', 'analyses', 'analysed', 'analysing', 'analysis',
    'realise', 'realises', 'realised', 'realising', 'realisation',
    'aluminium', 'honour', 'honours', 'honoured', 'honouring',
    'flavour', 'flavours', 'flavoured', 'flavouring',
    'neighbour', 'neighbours', 'neighbourhood',
    'defence', 'offence', 'licence', 'practise'
];
/**
 * File content samples for different programming languages
 */
exports.fileContents = {
    javascript: {
        withComments: `/**
 * Color analyzer utility
 */
function analyzeColor(color) {
    // Organize the color data
    const center = findColorCenter(color);
    return center;
}

// Export the analyzer
module.exports = { analyzeColor };`,
        withStrings: `const messages = {
    error: "Failed to analyze the color",
    success: "Color organization complete"
};

function processColor() {
    console.log("Processing color data");
    return "center";
}`
    },
    typescript: {
        withTypes: `interface ColorData {
    // Color properties
    value: string;
    center: Point;
}

/**
 * Analyzes color organization
 */
class ColorAnalyzer {
    // Organize the data
    analyze(color: ColorData): Point {
        return color.center;
    }
}`
    },
    python: {
        withDocstrings: `"""Color analysis module.

This module provides utilities for analyzing color organization.
"""

def analyze_color(color_value):
    """Analyze the color data.
    
    Args:
        color_value: The color to analyze
        
    Returns:
        The center point of the color
    """
    # Organize the data
    return find_center(color_value)`,
        withComments: `# Color analyzer
class ColorAnalyzer:
    def __init__(self):
        # Initialize the analyzer
        self.center = None
    
    def organize_data(self, data):
        # Organize the color data
        return sorted(data)`
    },
    go: {
        withComments: `// Package color provides color analysis utilities
package color

// ColorAnalyzer analyzes color organization
type ColorAnalyzer struct {
    // Center represents the color center
    Center Point
}

// AnalyzeColor analyzes the given color
func (ca *ColorAnalyzer) AnalyzeColor(color string) Point {
    // Organize the color data
    return ca.findCenter(color)
}`
    },
    markdown: {
        documentation: `# Color Analysis Guide

This document describes how to analyze color organization in your application.

## Getting Started

The color analyzer helps organize your color data:

1. Initialize the analyzer
2. Process your color data
3. Analyze the results

## Configuration

Configure the analyzer to center the color data properly.`,
        readme: `# Project Name

A color organization tool for analyzing data.

## Features

- Color analysis
- Data organization  
- Center point calculation

## Installation

\`\`\`bash
npm install color-analyzer
\`\`\`

## Usage

\`\`\`javascript
// Analyze colors
const analyzer = new ColorAnalyzer();
const result = analyzer.organize(colorData);
\`\`\``
    }
};
/**
 * Expected diagnostic ranges for test texts
 */
exports.expectedDiagnostics = {
    simple: [
        { word: 'color', line: 0, startChar: 4, endChar: 9 }
    ],
    multiple: [
        { word: 'color', line: 0, startChar: 4, endChar: 9 },
        { word: 'organization', line: 0, startChar: 17, endChar: 29 },
        { word: 'center', line: 0, startChar: 30, endChar: 36 },
        { word: 'analyzed', line: 0, startChar: 41, endChar: 49 }
    ],
    jsCode: [
        { word: 'color', line: 0, startChar: 7, endChar: 12 },
        { word: 'Organize', line: 2, startChar: 7, endChar: 15 }
    ]
};
/**
 * Unit conversion test cases
 */
exports.unitConversions = {
    temperature: {
        input: 'The room is 70°F',
        output: 'The room is 21°C',
        conversions: 1
    },
    distance: {
        input: 'The table is 5 feet long',
        output: 'The table is 1.52m long',
        conversions: 1
    },
    weight: {
        input: 'The package weighs 10 pounds',
        output: 'The package weighs 4.54kg',
        conversions: 1
    },
    multiple: {
        input: 'The room is 70°F and measures 10 feet by 12 feet, weighing 500 pounds',
        output: 'The room is 21°C and measures 3.05m by 3.66m, weighing 226.8kg',
        conversions: 4
    }
};
//# sourceMappingURL=testTexts.js.map