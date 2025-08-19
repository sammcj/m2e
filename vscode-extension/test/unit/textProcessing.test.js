"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
const assert = __importStar(require("assert"));
const textProcessing_1 = require("../../src/utils/textProcessing");
const vscode = __importStar(require("vscode"));
suite('Text Processing Test Suite', () => {
    suite('extractWordAtPosition', () => {
        test('should extract word at beginning of line', () => {
            const text = 'color is great';
            const position = new vscode.Position(0, 2); // Position in 'color'
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, 'color');
            assert.strictEqual(result.range.start.character, 0);
            assert.strictEqual(result.range.end.character, 5);
        });
        test('should extract word in middle of line', () => {
            const text = 'The color is great';
            const position = new vscode.Position(0, 6); // Position in 'color'
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, 'color');
            assert.strictEqual(result.range.start.character, 4);
            assert.strictEqual(result.range.end.character, 9);
        });
        test('should extract word at end of line', () => {
            const text = 'Great color';
            const position = new vscode.Position(0, 8); // Position in 'color'
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, 'color');
            assert.strictEqual(result.range.start.character, 6);
            assert.strictEqual(result.range.end.character, 11);
        });
        test('should handle multiline text', () => {
            const text = 'Line 1\nThe color is great\nLine 3';
            const position = new vscode.Position(1, 6); // Position in 'color' on line 2
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
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
            const result1 = (0, textProcessing_1.extractWordAtPosition)(text, position1);
            const result2 = (0, textProcessing_1.extractWordAtPosition)(text, position2);
            assert.strictEqual(result1.word, 'color');
            assert.strictEqual(result2.word, 'organization');
        });
        test('should handle empty or whitespace positions', () => {
            const text = 'word   word';
            const position = new vscode.Position(0, 5); // In whitespace
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, '');
        });
        test('should handle position out of bounds', () => {
            const text = 'short';
            const position = new vscode.Position(0, 100); // Beyond text length
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, '');
        });
        test('should handle hyphenated words', () => {
            const text = 'well-organized content';
            const position = new vscode.Position(0, 8); // In 'organized'
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, 'organized');
        });
    });
    suite('isAmericanSpelling', () => {
        test('should identify common American spellings', () => {
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('color'), true);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('organize'), true);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('center'), true);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('analyze'), true);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('realize'), true);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('aluminum'), true);
        });
        test('should not identify British spellings as American', () => {
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('colour'), false);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('organise'), false);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('centre'), false);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('analyse'), false);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('realise'), false);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('aluminium'), false);
        });
        test('should handle case variations', () => {
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('Color'), true);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('COLOR'), true);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('cOlOr'), true);
        });
        test('should handle empty or invalid words', () => {
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)(''), false);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('123'), false);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('!@#'), false);
        });
        test('should not match partial words', () => {
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('colors'), true); // Plural
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('coloring'), true); // Gerund
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('colorful'), true); // Compound
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('discolor'), true); // Prefix
        });
        test('should handle words that are not in dictionary', () => {
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('nonexistentword'), false);
            assert.strictEqual((0, textProcessing_1.isAmericanSpelling)('randomtext'), false);
        });
    });
    suite('shouldExcludeFromDiagnostics', () => {
        test('should exclude common file patterns', () => {
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('node_modules/package.json'), true);
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('.git/config'), true);
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('dist/bundle.js'), true);
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('build/output.js'), true);
        });
        test('should not exclude regular source files', () => {
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('src/index.js'), false);
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('lib/utils.ts'), false);
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('README.md'), false);
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('package.json'), false);
        });
        test('should handle custom exclude patterns', () => {
            const customPatterns = ['**/test/**', '**/docs/**'];
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('test/unit/test.js', customPatterns), true);
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('docs/api.md', customPatterns), true);
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('src/main.js', customPatterns), false);
        });
        test('should handle empty patterns', () => {
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)('any/file.js', []), false);
        });
        test('should handle absolute paths', () => {
            const absolutePath = '/Users/test/project/node_modules/package.json';
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)(absolutePath), true);
        });
        test('should handle Windows-style paths', () => {
            const windowsPath = 'C:\\project\\node_modules\\package.json';
            assert.strictEqual((0, textProcessing_1.shouldExcludeFromDiagnostics)(windowsPath), true);
        });
    });
    suite('Text processing edge cases', () => {
        test('should handle Unicode characters', () => {
            const text = 'The cölor is great'; // With umlaut
            const position = new vscode.Position(0, 6);
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, 'cölor');
        });
        test('should handle very long words', () => {
            const longWord = 'a'.repeat(1000);
            const text = `prefix ${longWord} suffix`;
            const position = new vscode.Position(0, 500); // Middle of long word
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, longWord);
        });
        test('should handle words with numbers', () => {
            const text = 'color2 and color3';
            const position1 = new vscode.Position(0, 3); // In 'color2'
            const position2 = new vscode.Position(0, 13); // In 'color3'
            const result1 = (0, textProcessing_1.extractWordAtPosition)(text, position1);
            const result2 = (0, textProcessing_1.extractWordAtPosition)(text, position2);
            assert.strictEqual(result1.word, 'color2');
            assert.strictEqual(result2.word, 'color3');
        });
        test('should handle camelCase words', () => {
            const text = 'backgroundColor';
            const position = new vscode.Position(0, 8); // In middle
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, 'backgroundColor');
        });
        test('should handle snake_case words', () => {
            const text = 'background_color';
            const position = new vscode.Position(0, 12); // In 'color' part
            const result = (0, textProcessing_1.extractWordAtPosition)(text, position);
            assert.strictEqual(result.word, 'background_color');
        });
    });
});
//# sourceMappingURL=textProcessing.test.js.map