const assert = require('assert');

/**
 * Pure utility functions that don't depend on VSCode APIs
 * These can be tested in any Node.js environment
 */

// Test basic utility functions that would exist in a real implementation
describe('Pure Utility Functions', function() {

    describe('Basic Text Processing', function() {
        
        function countWords(text) {
            return text.trim().split(/\s+/).filter(word => word.length > 0).length;
        }

        function countLines(text) {
            return text.split('\n').length;
        }

        function formatFileSize(bytes) {
            if (bytes === 0) return '0 B';
            
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }

        function truncateText(text, maxLength = 100) {
            if (text.length <= maxLength) {
                return text;
            }
            return text.substring(0, maxLength) + '...';
        }

        it('countWords should work correctly', function() {
            assert.strictEqual(countWords('hello world'), 2);
            assert.strictEqual(countWords(''), 0);
            assert.strictEqual(countWords('   '), 0);
            assert.strictEqual(countWords('one two three'), 3);
        });

        it('countLines should work correctly', function() {
            assert.strictEqual(countLines('single line'), 1);
            assert.strictEqual(countLines('line1\nline2'), 2);
            assert.strictEqual(countLines('line1\nline2\nline3'), 3);
        });

        it('formatFileSize should work correctly', function() {
            assert.strictEqual(formatFileSize(0), '0 B');
            assert.strictEqual(formatFileSize(1024), '1 KB');
            assert.strictEqual(formatFileSize(1048576), '1 MB');
        });

        it('truncateText should work correctly', function() {
            const longText = 'This is a very long text that should be truncated';
            const result = truncateText(longText, 20);
            assert.strictEqual(result, 'This is a very long ...');
            
            const shortText = 'Short';
            assert.strictEqual(truncateText(shortText, 20), 'Short');
        });
    });

    describe('Code Detection', function() {
        
        function looksLikeCode(text) {
            const codePatterns = [
                /\bfunction\s+\w+\s*\(/,
                /\bclass\s+\w+/,
                /\bimport\s+.+\bfrom\b/,
                /\bdef\s+\w+\s*\(/,
                /\bconsole\.log\(/,
                /\/\/.*$/m,
                /\/\*[\s\S]*?\*\//,
                /^<\w+.*>/m,
            ];

            return codePatterns.some(pattern => pattern.test(text));
        }

        function detectLanguage(text) {
            const languagePatterns = {
                javascript: [
                    /\bconsole\.log\(/,
                    /\bfunction\s+\w+\s*\(/,
                    /\bconst\s+\w+\s*=/,
                ],
                python: [
                    /\bdef\s+\w+\s*\(/,
                    /\bimport\s+\w+/,
                    /\bprint\s*\(/,
                ],
                go: [
                    /\bpackage\s+\w+/,
                    /\bfunc\s+\w+\s*\(/,
                    /\bfmt\.Print/,
                ],
                html: [
                    /^<!DOCTYPE/i,
                    /<html\b/i,
                    /<div\b/i,
                ],
            };

            for (const [language, patterns] of Object.entries(languagePatterns)) {
                if (patterns.some(pattern => pattern.test(text))) {
                    return language;
                }
            }

            return undefined;
        }

        it('looksLikeCode should identify code patterns', function() {
            assert.strictEqual(looksLikeCode('function it() { return true; }'), true);
            assert.strictEqual(looksLikeCode('console.log("hello");'), true);
            assert.strictEqual(looksLikeCode('def it(): pass'), true);
            assert.strictEqual(looksLikeCode('This is just text'), false);
        });

        it('detectLanguage should identify languages', function() {
            assert.strictEqual(detectLanguage('console.log("it");'), 'javascript');
            assert.strictEqual(detectLanguage('def it(): pass'), 'python');
            assert.strictEqual(detectLanguage('package main'), 'go');
            assert.strictEqual(detectLanguage('<!DOCTYPE html>'), 'html');
            assert.strictEqual(detectLanguage('regular text'), undefined);
        });
    });

    describe('Text Chunking', function() {
        
        function splitTextIntoChunks(text, maxChunkSize = 50000) {
            if (text.length <= maxChunkSize) {
                return [text];
            }

            const chunks = [];
            let currentIndex = 0;

            while (currentIndex < text.length) {
                let chunkEnd = currentIndex + maxChunkSize;
                
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

        function mergeChunks(chunks) {
            return chunks.join('');
        }

        it('splitTextIntoChunks should work correctly', function() {
            const shortText = 'Short text';
            assert.strictEqual(splitTextIntoChunks(shortText, 100).length, 1);

            const longText = 'a'.repeat(100);
            const chunks = splitTextIntoChunks(longText, 30);
            assert.strictEqual(chunks.length, 4);
            assert.strictEqual(chunks[0].length, 30);
        });

        it('mergeChunks should work correctly', function() {
            const chunks = ['chunk1', 'chunk2', 'chunk3'];
            const merged = mergeChunks(chunks);
            assert.strictEqual(merged, 'chunk1chunk2chunk3');
        });
    });

    describe('String Utilities', function() {
        
        function escapeMarkdown(text) {
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

        function calculateLevenshteinDistance(str1, str2) {
            const matrix = [];

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
                            matrix[i - 1][j - 1] + 1,
                            matrix[i][j - 1] + 1,
                            matrix[i - 1][j] + 1
                        );
                    }
                }
            }

            return matrix[str2.length][str1.length];
        }

        function calculateSimilarity(text1, text2) {
            if (text1 === text2) return 1.0;
            if (text1.length === 0 && text2.length === 0) return 1.0;
            if (text1.length === 0 || text2.length === 0) return 0.0;

            const longer = text1.length > text2.length ? text1 : text2;
            const shorter = text1.length > text2.length ? text2 : text1;
            
            if (longer.length === 0) return 1.0;

            const editDistance = calculateLevenshteinDistance(longer, shorter);
            return (longer.length - editDistance) / longer.length;
        }

        it('escapeMarkdown should escape special characters', function() {
            const text = '*bold* _italic_ `code`';
            const result = escapeMarkdown(text);
            assert.strictEqual(result, '\\*bold\\* \\_italic\\_ \\`code\\`');
        });

        it('calculateSimilarity should work correctly', function() {
            assert.strictEqual(calculateSimilarity('hello', 'hello'), 1.0);
            assert.strictEqual(calculateSimilarity('', ''), 1.0);
            assert.strictEqual(calculateSimilarity('hello', ''), 0.0);
            
            const similarity = calculateSimilarity('color', 'colour');
            assert.strictEqual(similarity > 0.5, true);
            assert.strictEqual(similarity < 1.0, true);
        });
    });
});