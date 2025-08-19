import * as vscode from 'vscode';
import { ConvertResponse } from '../services/client';

/**
 * Utility functions for text processing and manipulation
 */
export class TextProcessingUtils {
    
    /**
     * Validate text size and warn user if large
     */
    static async validateTextSize(text: string, operation: string, maxSize: number = 500000): Promise<boolean> {
        if (text.length > maxSize) {
            const sizeKB = Math.round(text.length / 1024);
            const proceed = await vscode.window.showWarningMessage(
                `M2E: Large text detected (${sizeKB}KB). ${operation} may take some time.`,
                'Continue',
                'Cancel'
            );
            return proceed === 'Continue';
        }
        return true;
    }

    /**
     * Count words in text for statistics
     */
    static countWords(text: string): number {
        return text.trim().split(/\s+/).filter(word => word.length > 0).length;
    }

    /**
     * Count lines in text
     */
    static countLines(text: string): number {
        return text.split('\n').length;
    }

    /**
     * Extract basic text statistics
     */
    static getTextStatistics(text: string): {
        characters: number;
        words: number;
        lines: number;
        size: string;
    } {
        const characters = text.length;
        const words = this.countWords(text);
        const lines = this.countLines(text);
        const size = this.formatFileSize(characters);

        return { characters, words, lines, size };
    }

    /**
     * Format file size in human-readable format
     */
    static formatFileSize(bytes: number): string {
        if (bytes === 0) return '0 B';
        
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    }

    /**
     * Truncate text for display purposes
     */
    static truncateText(text: string, maxLength: number = 100): string {
        if (text.length <= maxLength) {
            return text;
        }
        return text.substring(0, maxLength) + '...';
    }

    /**
     * Escape text for safe display in markdown
     */
    static escapeMarkdown(text: string): string {
        return text
            .replace(/\\/g, '\\\\')
            .replace(/\*/g, '\\*')
            .replace(/_/g, '\\_')
            .replace(/`/g, '\\`')
            .replace(/\[/g, '\\[')
            .replace(/\]/g, '\\]')
            .replace(/\(/g, '\\(')
            .replace(/\)/g, '\\)')
            .replace(/#/g, '\\#')
            .replace(/\+/g, '\\+')
            .replace(/-/g, '\\-')
            .replace(/\./g, '\\.')
            .replace(/!/g, '\\!');
    }

    /**
     * Generate a summary of conversion changes
     */
    static generateChangeSummary(result: ConvertResponse): string {
        const { metadata } = result;
        const totalChanges = metadata.spellingChanges + metadata.unitChanges;
        
        if (totalChanges === 0) {
            return 'No changes were made to the text.';
        }

        const parts: string[] = [];
        
        if (metadata.spellingChanges > 0) {
            parts.push(`${metadata.spellingChanges} spelling change${metadata.spellingChanges !== 1 ? 's' : ''}`);
        }
        
        if (metadata.unitChanges > 0) {
            parts.push(`${metadata.unitChanges} unit conversion${metadata.unitChanges !== 1 ? 's' : ''}`);
        }

        const summary = `Made ${totalChanges} total change${totalChanges !== 1 ? 's' : ''}: ${parts.join(' and ')}.`;
        
        if (metadata.processingTimeMs) {
            return `${summary} Processing took ${metadata.processingTimeMs}ms.`;
        }
        
        return summary;
    }

    /**
     * Check if text appears to be code based on common patterns
     */
    static looksLikeCode(text: string): boolean {
        const codePatterns = [
            /\bfunction\s+\w+\s*\(/,
            /\bclass\s+\w+/,
            /\bimport\s+.+\bfrom\b/,
            /\#include\s*<.+>/,
            /\bdef\s+\w+\s*\(/,
            /\bif\s*\(.+\)\s*\{/,
            /\bfor\s*\(.+\)\s*\{/,
            /\bwhile\s*\(.+\)\s*\{/,
            /\bconsole\.log\(/,
            /\bprint\(/,
            /\breturn\s+/,
            /\/\*[\s\S]*?\*\//, // Block comments
            /\/\/.*$/, // Line comments
            /^\s*<\w+.*>/m, // HTML/XML tags
        ];

        return codePatterns.some(pattern => pattern.test(text));
    }

    /**
     * Detect likely programming language from text content
     */
    static detectLanguage(text: string): string | undefined {
        const languagePatterns: { [key: string]: RegExp[] } = {
            javascript: [
                /\bconsole\.log\(/,
                /\bfunction\s+\w+\s*\(/,
                /\bconst\s+\w+\s*=/,
                /\blet\s+\w+\s*=/,
                /\bvar\s+\w+\s*=/,
                /\brequire\s*\(/,
                /\bmodule\.exports/
            ],
            typescript: [
                /\binterface\s+\w+/,
                /\btype\s+\w+\s*=/,
                /:\s*string\b/,
                /:\s*number\b/,
                /:\s*boolean\b/
            ],
            python: [
                /\bdef\s+\w+\s*\(/,
                /\bimport\s+\w+/,
                /\bfrom\s+\w+\s+import/,
                /\bprint\s*\(/,
                /\bif\s+__name__\s*==\s*['"']__main__['"']/
            ],
            java: [
                /\bpublic\s+class\s+\w+/,
                /\bpublic\s+static\s+void\s+main/,
                /\bSystem\.out\.println/,
                /\bpublic\s+\w+\s+\w+\s*\(/
            ],
            go: [
                /\bpackage\s+\w+/,
                /\bfunc\s+\w+\s*\(/,
                /\bimport\s*\(/,
                /\bfmt\.Print/,
                /\bvar\s+\w+\s+\w+/
            ],
            html: [
                /^\s*<!DOCTYPE/i,
                /<html\b/i,
                /<head\b/i,
                /<body\b/i,
                /<div\b/i
            ],
            css: [
                /^\s*[\w\-\.#]+\s*\{/m,
                /\w+\s*:\s*[\w\-#%]+\s*;/,
                /@media\s+/,
                /@import\s+/
            ]
        };

        for (const [language, patterns] of Object.entries(languagePatterns)) {
            if (patterns.some(pattern => pattern.test(text))) {
                return language;
            }
        }

        return undefined;
    }

    /**
     * Split text into chunks for processing large files
     */
    static splitTextIntoChunks(text: string, maxChunkSize: number = 50000): string[] {
        if (text.length <= maxChunkSize) {
            return [text];
        }

        const chunks: string[] = [];
        let currentIndex = 0;

        while (currentIndex < text.length) {
            let chunkEnd = currentIndex + maxChunkSize;
            
            // Try to break at word boundaries to avoid splitting words
            if (chunkEnd < text.length) {
                const nextSpace = text.lastIndexOf(' ', chunkEnd);
                const nextNewline = text.lastIndexOf('\n', chunkEnd);
                const breakPoint = Math.max(nextSpace, nextNewline);
                
                if (breakPoint > currentIndex) {
                    chunkEnd = breakPoint + 1;
                }
            }

            chunks.push(text.substring(currentIndex, chunkEnd));
            currentIndex = chunkEnd;
        }

        return chunks;
    }

    /**
     * Merge processed chunks back together
     */
    static mergeChunks(chunks: string[]): string {
        return chunks.join('');
    }

    /**
     * Calculate similarity between two texts (simple implementation)
     */
    static calculateSimilarity(text1: string, text2: string): number {
        if (text1 === text2) return 1.0;
        if (text1.length === 0 && text2.length === 0) return 1.0;
        if (text1.length === 0 || text2.length === 0) return 0.0;

        // Simple character-based similarity
        const longer = text1.length > text2.length ? text1 : text2;
        const shorter = text1.length > text2.length ? text2 : text1;
        
        if (longer.length === 0) return 1.0;

        const editDistance = this.levenshteinDistance(longer, shorter);
        return (longer.length - editDistance) / longer.length;
    }

    /**
     * Calculate Levenshtein distance between two strings
     */
    private static levenshteinDistance(str1: string, str2: string): number {
        const matrix: number[][] = [];

        for (let i = 0; i <= str2.length; i++) {
            matrix[i] = [i];
        }

        for (let j = 0; j <= str1.length; j++) {
            matrix[0][j] = j;
        }

        for (let i = 1; i <= str2.length; i++) {
            for (let j = 1; j <= str1.length; j++) {
                if (str2.charAt(i - 1) === str1.charAt(j - 1)) {
                    matrix[i][j] = matrix[i - 1][j - 1];
                } else {
                    matrix[i][j] = Math.min(
                        matrix[i - 1][j - 1] + 1, // substitution
                        matrix[i][j - 1] + 1,     // insertion
                        matrix[i - 1][j] + 1      // deletion
                    );
                }
            }
        }

        return matrix[str2.length][str1.length];
    }
}

/**
 * Diff utility functions for comparing texts
 */
export class DiffUtils {
    
    /**
     * Generate a simple diff report
     */
    static generateDiffReport(original: string, converted: string, changes: ConvertResponse['changes']): string {
        const lines: string[] = [
            '# M2E Conversion Diff Report',
            '',
            `**Original Length**: ${original.length} characters`,
            `**Converted Length**: ${converted.length} characters`,
            `**Changes Made**: ${changes.length}`,
            ''
        ];

        if (changes.length > 0) {
            lines.push('## Changes Detail', '');
            
            changes.forEach((change, index) => {
                const context = this.getChangeContext(original, change.position, change.original.length);
                lines.push(`### Change ${index + 1} (${change.type})`);
                lines.push(`- **Position**: ${change.position}`);
                lines.push(`- **Original**: \`${change.original}\``);
                lines.push(`- **Converted**: \`${change.converted}\``);
                lines.push(`- **Context**: ...${context}...`);
                lines.push('');
            });
        } else {
            lines.push('## No Changes', '', 'The text did not require any conversions.');
        }

        return lines.join('\n');
    }

    /**
     * Get context around a change for better understanding
     */
    private static getChangeContext(text: string, position: number, length: number, contextSize: number = 20): string {
        const start = Math.max(0, position - contextSize);
        const end = Math.min(text.length, position + length + contextSize);
        
        return text.substring(start, end)
            .replace(/\n/g, '\\n')
            .replace(/\t/g, '\\t');
    }

    /**
     * Highlight changes in text for display
     */
    static highlightChanges(text: string, changes: ConvertResponse['changes']): string {
        let highlightedText = text;
        let offset = 0;

        // Sort changes by position to apply them in order
        const sortedChanges = [...changes].sort((a, b) => a.position - b.position);

        for (const change of sortedChanges) {
            const position = change.position + offset;
            const before = highlightedText.substring(0, position);
            const after = highlightedText.substring(position + change.original.length);
            const highlighted = `**${change.converted}**`;
            
            highlightedText = before + highlighted + after;
            offset += highlighted.length - change.original.length;
        }

        return highlightedText;
    }

    /**
     * Create ranges for VSCode editor highlighting
     */
    static createHighlightRanges(document: vscode.TextDocument, changes: ConvertResponse['changes']): {
        ranges: vscode.Range[];
        decorations: vscode.DecorationOptions[];
    } {
        const ranges: vscode.Range[] = [];
        const decorations: vscode.DecorationOptions[] = [];

        for (const change of changes) {
            try {
                const startPos = document.positionAt(change.position);
                const endPos = document.positionAt(change.position + change.original.length);
                const range = new vscode.Range(startPos, endPos);
                
                ranges.push(range);
                decorations.push({
                    range,
                    hoverMessage: `M2E: "${change.original}" → "${change.converted}" (${change.type})`
                });
            } catch (error) {
                // Skip invalid positions
                console.warn('Invalid position in change:', change);
            }
        }

        return { ranges, decorations };
    }
}