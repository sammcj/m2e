/**
 * Sample file contents for testing M2E extension with different file types
 */

export const sampleFiles = {
    /**
     * JavaScript files
     */
    javascript: {
        simple: `// Color analyzer utility
function analyzeColor(color) {
    return color;
}`,

        complex: `/**
 * Color analysis module
 * Provides utilities for analyzing color organization
 */

// Color analyzer class
class ColorAnalyzer {
    constructor() {
        // Initialize the color center
        this.center = { x: 0, y: 0 };
    }

    /**
     * Analyze the color organization
     * @param {string} color - The color to analyze
     * @returns {Object} Analysis results
     */
    analyzeColor(color) {
        // Organize the color data
        const organized = this.organizeData(color);
        
        // Find the center point
        const center = this.findCenter(organized);
        
        return {
            color: color,
            center: center,
            organized: organized
        };
    }

    // Helper method to organize data
    organizeData(data) {
        // Simple organization algorithm
        return data.sort();
    }

    // Find the center of the color space
    findCenter(data) {
        // Calculate center coordinates
        return this.center;
    }
}

module.exports = { ColorAnalyzer };`,

        withStrings: `const config = {
    messages: {
        // Error messages
        colorError: "Failed to analyze the color",
        organizationError: "Unable to organize the data",
        centerError: "Cannot find the center point"
    },
    
    defaults: {
        color: "red",
        organization: "grid",
        center: { x: 50, y: 50 }
    }
};

// Process color configuration
function processConfig() {
    console.log("Processing color organization config");
    console.log("Center point:", config.defaults.center);
}`
    },

    /**
     * TypeScript files
     */
    typescript: {
        interfaces: `/**
 * Color analysis interfaces
 */

// Color data interface
interface ColorData {
    value: string;
    organization: 'grid' | 'radial' | 'linear';
    center: Point;
}

// Point interface for center coordinates
interface Point {
    x: number;
    y: number;
}

// Analysis result interface
interface AnalysisResult {
    originalColor: ColorData;
    analyzedColor: ColorData;
    organization: string;
    processingTime: number;
}`,

        classes: `import { ColorData, Point, AnalysisResult } from './interfaces';

/**
 * Advanced color analyzer with TypeScript support
 */
export class ColorAnalyzer {
    private center: Point;
    private organization: string;

    constructor() {
        // Initialize analyzer with default center
        this.center = { x: 0, y: 0 };
        this.organization = 'grid';
    }

    /**
     * Analyze color data with proper organization
     */
    public analyzeColor(data: ColorData): AnalysisResult {
        // Organize the input data
        const organized = this.organizeColorData(data);
        
        // Find the optimal center
        const newCenter = this.calculateCenter(organized);
        
        return {
            originalColor: data,
            analyzedColor: organized,
            organization: this.organization,
            processingTime: Date.now()
        };
    }

    // Private method to organize color data
    private organizeColorData(data: ColorData): ColorData {
        // Advanced organization algorithm
        return {
            ...data,
            organization: this.organization,
            center: this.center
        };
    }

    // Calculate the optimal center point
    private calculateCenter(data: ColorData): Point {
        // Center calculation logic
        return this.center;
    }
}`
    },

    /**
     * Python files
     */
    python: {
        simple: `# Color analyzer module
def analyze_color(color):
    """Analyze the color data."""
    # Simple color analysis
    return color`,

        classes: `"""Color analysis module.

This module provides utilities for analyzing color organization
and finding the center point of color data.
"""

class ColorAnalyzer:
    """Analyzes color organization and center points."""
    
    def __init__(self):
        """Initialize the color analyzer."""
        # Set default center point
        self.center = (0, 0)
        self.organization = 'grid'
    
    def analyze_color(self, color_data):
        """Analyze the color data.
        
        Args:
            color_data: The color information to analyze
            
        Returns:
            dict: Analysis results including organization and center
        """
        # Organize the color data
        organized_data = self._organize_data(color_data)
        
        # Find the center point
        center_point = self._find_center(organized_data)
        
        return {
            'color': color_data,
            'organization': self.organization,
            'center': center_point
        }
    
    def _organize_data(self, data):
        """Organize the color data efficiently."""
        # Private method for data organization
        return sorted(data) if isinstance(data, list) else data
    
    def _find_center(self, data):
        """Find the center point of the organized data."""
        # Center finding algorithm
        return self.center`,

        withComments: `#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# Color analysis utility
# Author: Developer
# Purpose: Analyze color organization

import math
import statistics

# Global configuration
DEFAULT_COLOR = 'blue'
DEFAULT_ORGANIZATION = 'center'

def main():
    """Main function to demonstrate color analysis."""
    # Initialize the color analyzer
    analyzer = ColorAnalyzer()
    
    # Test data for color analysis
    test_colors = ['red', 'green', 'blue']
    
    # Analyze each color
    for color in test_colors:
        # Process the color
        result = analyzer.analyze_color(color)
        print(f"Color: {color}, Center: {result['center']}")

if __name__ == '__main__':
    main()`
    },

    /**
     * Go files
     */
    go: {
        package: `// Package color provides color analysis utilities
package color

import (
    "fmt"
    "sort"
)

// ColorData represents color information
type ColorData struct {
    Value        string  // Color value
    Organization string  // Organization type
    Center       Point   // Center coordinates
}

// Point represents a 2D coordinate
type Point struct {
    X float64 // X coordinate
    Y float64 // Y coordinate
}

// ColorAnalyzer analyzes color organization
type ColorAnalyzer struct {
    center       Point  // Default center point
    organization string // Organization method
}

// NewColorAnalyzer creates a new color analyzer
func NewColorAnalyzer() *ColorAnalyzer {
    return &ColorAnalyzer{
        center:       Point{X: 0, Y: 0},
        organization: "grid",
    }
}

// AnalyzeColor analyzes the given color data
func (ca *ColorAnalyzer) AnalyzeColor(data ColorData) (ColorData, error) {
    // Organize the color data
    organized := ca.organizeData(data)
    
    // Calculate the center point
    center := ca.calculateCenter(organized)
    organized.Center = center
    
    return organized, nil
}

// organizeData organizes the color data efficiently
func (ca *ColorAnalyzer) organizeData(data ColorData) ColorData {
    // Simple organization logic
    data.Organization = ca.organization
    return data
}

// calculateCenter finds the center point
func (ca *ColorAnalyzer) calculateCenter(data ColorData) Point {
    // Center calculation algorithm
    return ca.center
}`
    },

    /**
     * Markdown files
     */
    markdown: {
        documentation: `# Color Analysis Documentation

This document describes the color analysis and organization system.

## Overview

The color analyzer provides utilities for:

- Color data analysis
- Organization of color information
- Center point calculation
- Data visualization

## Getting Started

### Installation

\`\`\`bash
npm install color-analyzer
\`\`\`

### Basic Usage

\`\`\`javascript
// Initialize the analyzer
const analyzer = new ColorAnalyzer();

// Analyze a color
const result = analyzer.analyzeColor({
    value: 'blue',
    organization: 'grid'
});
\`\`\`

## Configuration

The analyzer can be configured to use different organization methods:

- **Grid**: Organize colors in a grid pattern
- **Radial**: Organize colors in a radial pattern  
- **Linear**: Organize colors linearly

### Setting the Center Point

You can specify a custom center point for the analysis:

\`\`\`javascript
analyzer.setCenter({ x: 100, y: 100 });
\`\`\`

## API Reference

### ColorAnalyzer

Main class for color analysis and organization.

#### Methods

- \`analyzeColor(data)\` - Analyze color data
- \`organizeData(data)\` - Organize color information
- \`findCenter(data)\` - Calculate center point

## Examples

### Basic Color Analysis

\`\`\`javascript
const colorData = {
    value: 'red',
    organization: 'grid'
};

const result = analyzer.analyzeColor(colorData);
console.log('Center:', result.center);
\`\`\`

### Advanced Organization

\`\`\`javascript
// Set up custom organization
analyzer.setOrganization('radial');
analyzer.setCenter({ x: 50, y: 50 });

// Analyze multiple colors
const colors = ['red', 'green', 'blue'];
const results = colors.map(color => 
    analyzer.analyzeColor({ value: color })
);
\`\`\``,

        readme: `# Color Analyzer

A powerful utility for color analysis and organization.

## Features

- **Color Analysis**: Analyze color data efficiently
- **Data Organization**: Organize colors using various methods
- **Center Calculation**: Find optimal center points
- **Multiple Formats**: Support for various color formats

## Installation

\`\`\`bash
npm install color-analyzer
\`\`\`

## Quick Start

\`\`\`javascript
const { ColorAnalyzer } = require('color-analyzer');

// Create analyzer
const analyzer = new ColorAnalyzer();

// Analyze a color
const result = analyzer.analyzeColor('blue');
console.log('Analysis complete:', result);
\`\`\`

## Documentation

See the [full documentation](./docs/README.md) for detailed usage instructions.

## License

MIT - see [LICENSE](LICENSE) file for details.`
    },

    /**
     * Configuration files
     */
    json: {
        packageJson: `{
    "name": "color-analyzer",
    "version": "1.0.0",
    "description": "A utility for color analysis and organization",
    "main": "index.js",
    "scripts": {
        "test": "jest",
        "start": "node index.js",
        "analyze": "node scripts/analyze.js"
    },
    "keywords": [
        "color",
        "analysis",
        "organization",
        "center",
        "utility"
    ],
    "author": "Developer",
    "license": "MIT"
}`,

        config: `{
    "colorAnalysis": {
        "defaultOrganization": "grid",
        "centerCalculation": "automatic",
        "precision": 2
    },
    "display": {
        "showCenter": true,
        "highlightOrganization": true,
        "colorScheme": "default"
    },
    "performance": {
        "cacheResults": true,
        "batchSize": 100,
        "timeoutMs": 5000
    }
}`
    }
};

/**
 * Sample workspace configurations for testing
 */
export const workspaceConfigs = {
    default: {
        "m2e.enableDiagnostics": true,
        "m2e.diagnosticSeverity": "Information",
        "m2e.enableUnitConversion": true,
        "m2e.serverPort": 18181,
        "m2e.showStatusBar": true,
        "m2e.codeAwareConversion": true,
        "m2e.preserveCodeSyntax": true,
        "m2e.debugLogging": false
    },

    disabled: {
        "m2e.enableDiagnostics": false,
        "m2e.enableUnitConversion": false,
        "m2e.showStatusBar": false
    },

    development: {
        "m2e.enableDiagnostics": true,
        "m2e.diagnosticSeverity": "Warning",
        "m2e.debugLogging": true,
        "m2e.serverPort": 18182
    },

    production: {
        "m2e.enableDiagnostics": true,
        "m2e.diagnosticSeverity": "Information",
        "m2e.debugLogging": false,
        "m2e.excludePatterns": [
            "**/node_modules/**",
            "**/.git/**",
            "**/dist/**",
            "**/build/**",
            "**/coverage/**"
        ]
    }
};

/**
 * Expected conversion results for sample files
 */
export const expectedConversions = {
    javascript: {
        simple: {
            spellingChanges: 2, // "Color" and "analyzeColor"
            lines: [0, 1]
        },
        complex: {
            spellingChanges: 8, // Multiple instances of color/analyze/organize
            commentLines: [0, 1, 5, 19, 29, 37, 43, 49]
        }
    },

    python: {
        classes: {
            spellingChanges: 6, // color/analyze/organize instances
            docstringLines: [0, 7, 16],
            commentLines: [9, 37, 44]
        }
    },

    markdown: {
        documentation: {
            spellingChanges: 12, // Multiple instances throughout
            affectedSections: ['overview', 'configuration', 'examples']
        }
    }
};