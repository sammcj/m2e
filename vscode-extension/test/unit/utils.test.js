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
const vscode = __importStar(require("vscode"));
const utils_1 = require("../../src/utils");
suite('Utils Test Suite', () => {
    suite('isTextFile', () => {
        test('should identify text files correctly', () => {
            assert.strictEqual((0, utils_1.isTextFile)('test.txt'), true);
            assert.strictEqual((0, utils_1.isTextFile)('test.md'), true);
            assert.strictEqual((0, utils_1.isTextFile)('test.js'), true);
            assert.strictEqual((0, utils_1.isTextFile)('test.ts'), true);
            assert.strictEqual((0, utils_1.isTextFile)('test.py'), true);
        });
        test('should reject binary files', () => {
            assert.strictEqual((0, utils_1.isTextFile)('test.png'), false);
            assert.strictEqual((0, utils_1.isTextFile)('test.jpg'), false);
            assert.strictEqual((0, utils_1.isTextFile)('test.pdf'), false);
            assert.strictEqual((0, utils_1.isTextFile)('test.zip'), false);
            assert.strictEqual((0, utils_1.isTextFile)('test.exe'), false);
        });
        test('should handle files without extensions', () => {
            assert.strictEqual((0, utils_1.isTextFile)('README'), true);
            assert.strictEqual((0, utils_1.isTextFile)('Makefile'), true);
            assert.strictEqual((0, utils_1.isTextFile)('LICENSE'), true);
        });
        test('should handle empty or invalid filenames', () => {
            assert.strictEqual((0, utils_1.isTextFile)(''), false);
            assert.strictEqual((0, utils_1.isTextFile)('.'), false);
            assert.strictEqual((0, utils_1.isTextFile)('..'), false);
        });
    });
    suite('getFileType', () => {
        test('should return correct file types for extensions', () => {
            assert.strictEqual((0, utils_1.getFileType)('test.js'), 'javascript');
            assert.strictEqual((0, utils_1.getFileType)('test.ts'), 'typescript');
            assert.strictEqual((0, utils_1.getFileType)('test.py'), 'python');
            assert.strictEqual((0, utils_1.getFileType)('test.go'), 'go');
            assert.strictEqual((0, utils_1.getFileType)('test.java'), 'java');
            assert.strictEqual((0, utils_1.getFileType)('test.md'), 'markdown');
            assert.strictEqual((0, utils_1.getFileType)('test.txt'), 'text');
        });
        test('should handle case insensitive extensions', () => {
            assert.strictEqual((0, utils_1.getFileType)('test.JS'), 'javascript');
            assert.strictEqual((0, utils_1.getFileType)('test.PY'), 'python');
            assert.strictEqual((0, utils_1.getFileType)('test.MD'), 'markdown');
        });
        test('should return text for unknown extensions', () => {
            assert.strictEqual((0, utils_1.getFileType)('test.xyz'), 'text');
            assert.strictEqual((0, utils_1.getFileType)('test.unknown'), 'text');
        });
        test('should handle files without extensions', () => {
            assert.strictEqual((0, utils_1.getFileType)('README'), 'text');
            assert.strictEqual((0, utils_1.getFileType)('Makefile'), 'text');
        });
    });
    suite('formatBytes', () => {
        test('should format bytes correctly', () => {
            assert.strictEqual((0, utils_1.formatBytes)(0), '0 Bytes');
            assert.strictEqual((0, utils_1.formatBytes)(1024), '1 KB');
            assert.strictEqual((0, utils_1.formatBytes)(1048576), '1 MB');
            assert.strictEqual((0, utils_1.formatBytes)(1073741824), '1 GB');
        });
        test('should handle decimal places', () => {
            assert.strictEqual((0, utils_1.formatBytes)(1536), '1.5 KB'); // 1.5 * 1024
            assert.strictEqual((0, utils_1.formatBytes)(1572864), '1.5 MB'); // 1.5 * 1024 * 1024
        });
        test('should handle negative numbers', () => {
            assert.strictEqual((0, utils_1.formatBytes)(-1024), '-1 KB');
        });
        test('should handle very large numbers', () => {
            const result = (0, utils_1.formatBytes)(1099511627776); // 1 TB
            assert.ok(result.includes('TB'));
        });
    });
    suite('debounce', () => {
        test('should delay function execution', (done) => {
            let callCount = 0;
            const debouncedFn = (0, utils_1.debounce)(() => {
                callCount++;
            }, 50);
            // Call multiple times rapidly
            debouncedFn();
            debouncedFn();
            debouncedFn();
            // Should not have been called yet
            assert.strictEqual(callCount, 0);
            // Wait for debounce delay
            setTimeout(() => {
                assert.strictEqual(callCount, 1);
                done();
            }, 100);
        });
        test('should pass arguments correctly', (done) => {
            let receivedArgs = [];
            const debouncedFn = (0, utils_1.debounce)((...args) => {
                receivedArgs = args;
            }, 50);
            debouncedFn('test', 123, true);
            setTimeout(() => {
                assert.deepStrictEqual(receivedArgs, ['test', 123, true]);
                done();
            }, 100);
        });
        test('should cancel previous calls', (done) => {
            let callCount = 0;
            const debouncedFn = (0, utils_1.debounce)(() => {
                callCount++;
            }, 50);
            debouncedFn();
            setTimeout(() => {
                debouncedFn(); // This should cancel the first call
            }, 25);
            setTimeout(() => {
                assert.strictEqual(callCount, 1); // Should only be called once
                done();
            }, 150);
        });
    });
    suite('isLargeSelection', () => {
        test('should identify large selections', () => {
            const smallText = 'a'.repeat(1000); // 1KB
            const mediumText = 'a'.repeat(50000); // 50KB  
            const largeText = 'a'.repeat(150000); // 150KB
            assert.strictEqual((0, utils_1.isLargeSelection)(smallText), false);
            assert.strictEqual((0, utils_1.isLargeSelection)(mediumText), false);
            assert.strictEqual((0, utils_1.isLargeSelection)(largeText), true);
        });
        test('should handle empty strings', () => {
            assert.strictEqual((0, utils_1.isLargeSelection)(''), false);
        });
        test('should use custom threshold', () => {
            const text = 'a'.repeat(5000);
            assert.strictEqual((0, utils_1.isLargeSelection)(text, 1000), true);
            assert.strictEqual((0, utils_1.isLargeSelection)(text, 10000), false);
        });
    });
    suite('VSCode API utilities', () => {
        test('should create position correctly', () => {
            const position = new vscode.Position(5, 10);
            assert.strictEqual(position.line, 5);
            assert.strictEqual(position.character, 10);
        });
        test('should create range correctly', () => {
            const start = new vscode.Position(0, 0);
            const end = new vscode.Position(5, 10);
            const range = new vscode.Range(start, end);
            assert.strictEqual(range.start.line, 0);
            assert.strictEqual(range.start.character, 0);
            assert.strictEqual(range.end.line, 5);
            assert.strictEqual(range.end.character, 10);
        });
        test('should create selection correctly', () => {
            const anchor = new vscode.Position(2, 5);
            const active = new vscode.Position(3, 10);
            const selection = new vscode.Selection(anchor, active);
            assert.strictEqual(selection.anchor.line, 2);
            assert.strictEqual(selection.anchor.character, 5);
            assert.strictEqual(selection.active.line, 3);
            assert.strictEqual(selection.active.character, 10);
        });
    });
});
//# sourceMappingURL=utils.test.js.map