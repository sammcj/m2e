import * as assert from 'assert';
import { 
    extractWordAtPosition, 
    isAmericanSpelling, 
    shouldExcludeFromDiagnostics 
} from '../../src/utils/textProcessing';
import * as vscode from 'vscode';

suite('Text Processing Test Suite', () => {
    
    suite('extractWordAtPosition', () => {
        test('should extract word at beginning of line', () => {
            const text = 'color is great';
            const position = new vscode.Position(0, 2); // Position in 'color'
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, 'color');
            assert.strictEqual(result.range.start.character, 0);
            assert.strictEqual(result.range.end.character, 5);
        });

        test('should extract word in middle of line', () => {
            const text = 'The color is great';
            const position = new vscode.Position(0, 6); // Position in 'color'
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, 'color');
            assert.strictEqual(result.range.start.character, 4);
            assert.strictEqual(result.range.end.character, 9);
        });

        test('should extract word at end of line', () => {
            const text = 'Great color';
            const position = new vscode.Position(0, 8); // Position in 'color'
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, 'color');
            assert.strictEqual(result.range.start.character, 6);
            assert.strictEqual(result.range.end.character, 11);
        });

        test('should handle multiline text', () => {
            const text = 'Line 1\nThe color is great\nLine 3';
            const position = new vscode.Position(1, 6); // Position in 'color' on line 2
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, 'color');
            assert.strictEqual(result.range.start.line, 1);
            assert.strictEqual(result.range.start.character, 4);
            assert.strictEqual(result.range.end.line, 1);
            assert.strictEqual(result.range.end.character, 9);
        });

        test('should handle punctuation boundaries', () => {
            const text = 'The color, organization.';
            const position1 = new vscode.Position(0, 6); // In 'color'
            const position2 = new vscode.Position(0, 15); // In 'organization'
            
            const result1 = extractWordAtPosition(text, position1);
            const result2 = extractWordAtPosition(text, position2);
            
            assert.strictEqual(result1.word, 'color');
            assert.strictEqual(result2.word, 'organization');
        });

        test('should handle empty or whitespace positions', () => {
            const text = 'word   word';
            const position = new vscode.Position(0, 5); // In whitespace
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, '');
        });

        test('should handle position out of bounds', () => {
            const text = 'short';
            const position = new vscode.Position(0, 100); // Beyond text length
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, '');
        });

        test('should handle hyphenated words', () => {
            const text = 'well-organized content';
            const position = new vscode.Position(0, 8); // In 'organized'
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, 'organized');
        });
    });

    suite('isAmericanSpelling', () => {
        test('should identify common American spellings', () => {
            assert.strictEqual(isAmericanSpelling('color'), true);
            assert.strictEqual(isAmericanSpelling('organize'), true);
            assert.strictEqual(isAmericanSpelling('center'), true);
            assert.strictEqual(isAmericanSpelling('analyze'), true);
            assert.strictEqual(isAmericanSpelling('realize'), true);
            assert.strictEqual(isAmericanSpelling('aluminum'), true);
        });

        test('should not identify British spellings as American', () => {
            assert.strictEqual(isAmericanSpelling('colour'), false);
            assert.strictEqual(isAmericanSpelling('organise'), false);
            assert.strictEqual(isAmericanSpelling('centre'), false);
            assert.strictEqual(isAmericanSpelling('analyse'), false);
            assert.strictEqual(isAmericanSpelling('realise'), false);
            assert.strictEqual(isAmericanSpelling('aluminium'), false);
        });

        test('should handle case variations', () => {
            assert.strictEqual(isAmericanSpelling('Color'), true);
            assert.strictEqual(isAmericanSpelling('COLOR'), true);
            assert.strictEqual(isAmericanSpelling('cOlOr'), true);
        });

        test('should handle empty or invalid words', () => {
            assert.strictEqual(isAmericanSpelling(''), false);
            assert.strictEqual(isAmericanSpelling('123'), false);
            assert.strictEqual(isAmericanSpelling('!@#'), false);
        });

        test('should not match partial words', () => {
            assert.strictEqual(isAmericanSpelling('colors'), true); // Plural
            assert.strictEqual(isAmericanSpelling('coloring'), true); // Gerund
            assert.strictEqual(isAmericanSpelling('colorful'), true); // Compound
            assert.strictEqual(isAmericanSpelling('discolor'), true); // Prefix
        });

        test('should handle words that are not in dictionary', () => {
            assert.strictEqual(isAmericanSpelling('nonexistentword'), false);
            assert.strictEqual(isAmericanSpelling('randomtext'), false);
        });
    });

    suite('shouldExcludeFromDiagnostics', () => {
        test('should exclude common file patterns', () => {
            assert.strictEqual(shouldExcludeFromDiagnostics('node_modules/package.json'), true);
            assert.strictEqual(shouldExcludeFromDiagnostics('.git/config'), true);
            assert.strictEqual(shouldExcludeFromDiagnostics('dist/bundle.js'), true);
            assert.strictEqual(shouldExcludeFromDiagnostics('build/output.js'), true);
        });

        test('should not exclude regular source files', () => {
            assert.strictEqual(shouldExcludeFromDiagnostics('src/index.js'), false);
            assert.strictEqual(shouldExcludeFromDiagnostics('lib/utils.ts'), false);
            assert.strictEqual(shouldExcludeFromDiagnostics('README.md'), false);
            assert.strictEqual(shouldExcludeFromDiagnostics('package.json'), false);
        });

        test('should handle custom exclude patterns', () => {
            const customPatterns = ['**/test/**', '**/docs/**'];
            
            assert.strictEqual(
                shouldExcludeFromDiagnostics('test/unit/test.js', customPatterns), 
                true
            );
            assert.strictEqual(
                shouldExcludeFromDiagnostics('docs/api.md', customPatterns), 
                true
            );
            assert.strictEqual(
                shouldExcludeFromDiagnostics('src/main.js', customPatterns), 
                false
            );
        });

        test('should handle empty patterns', () => {
            assert.strictEqual(shouldExcludeFromDiagnostics('any/file.js', []), false);
        });

        test('should handle absolute paths', () => {
            const absolutePath = '/Users/test/project/node_modules/package.json';
            assert.strictEqual(shouldExcludeFromDiagnostics(absolutePath), true);
        });

        test('should handle Windows-style paths', () => {
            const windowsPath = 'C:\\project\\node_modules\\package.json';
            assert.strictEqual(shouldExcludeFromDiagnostics(windowsPath), true);
        });
    });

    suite('Text processing edge cases', () => {
        test('should handle Unicode characters', () => {
            const text = 'The cölor is great'; // With umlaut
            const position = new vscode.Position(0, 6);
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, 'cölor');
        });

        test('should handle very long words', () => {
            const longWord = 'a'.repeat(1000);
            const text = `prefix ${longWord} suffix`;
            const position = new vscode.Position(0, 500); // Middle of long word
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, longWord);
        });

        test('should handle words with numbers', () => {
            const text = 'color2 and color3';
            const position1 = new vscode.Position(0, 3); // In 'color2'
            const position2 = new vscode.Position(0, 13); // In 'color3'
            
            const result1 = extractWordAtPosition(text, position1);
            const result2 = extractWordAtPosition(text, position2);
            
            assert.strictEqual(result1.word, 'color2');
            assert.strictEqual(result2.word, 'color3');
        });

        test('should handle camelCase words', () => {
            const text = 'backgroundColor';
            const position = new vscode.Position(0, 8); // In middle
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, 'backgroundColor');
        });

        test('should handle snake_case words', () => {
            const text = 'background_color';
            const position = new vscode.Position(0, 12); // In 'color' part
            const result = extractWordAtPosition(text, position);
            
            assert.strictEqual(result.word, 'background_color');
        });
    });
});