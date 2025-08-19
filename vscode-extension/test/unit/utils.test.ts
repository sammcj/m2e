import * as assert from 'assert';
import * as vscode from 'vscode';
import { 
    isTextFile, 
    getFileType, 
    formatBytes, 
    debounce, 
    isLargeSelection 
} from '../../src/utils';

suite('Utils Test Suite', () => {
    
    suite('isTextFile', () => {
        test('should identify text files correctly', () => {
            assert.strictEqual(isTextFile('test.txt'), true);
            assert.strictEqual(isTextFile('test.md'), true);
            assert.strictEqual(isTextFile('test.js'), true);
            assert.strictEqual(isTextFile('test.ts'), true);
            assert.strictEqual(isTextFile('test.py'), true);
        });

        test('should reject binary files', () => {
            assert.strictEqual(isTextFile('test.png'), false);
            assert.strictEqual(isTextFile('test.jpg'), false);
            assert.strictEqual(isTextFile('test.pdf'), false);
            assert.strictEqual(isTextFile('test.zip'), false);
            assert.strictEqual(isTextFile('test.exe'), false);
        });

        test('should handle files without extensions', () => {
            assert.strictEqual(isTextFile('README'), true);
            assert.strictEqual(isTextFile('Makefile'), true);
            assert.strictEqual(isTextFile('LICENSE'), true);
        });

        test('should handle empty or invalid filenames', () => {
            assert.strictEqual(isTextFile(''), false);
            assert.strictEqual(isTextFile('.'), false);
            assert.strictEqual(isTextFile('..'), false);
        });
    });

    suite('getFileType', () => {
        test('should return correct file types for extensions', () => {
            assert.strictEqual(getFileType('test.js'), 'javascript');
            assert.strictEqual(getFileType('test.ts'), 'typescript');
            assert.strictEqual(getFileType('test.py'), 'python');
            assert.strictEqual(getFileType('test.go'), 'go');
            assert.strictEqual(getFileType('test.java'), 'java');
            assert.strictEqual(getFileType('test.md'), 'markdown');
            assert.strictEqual(getFileType('test.txt'), 'text');
        });

        test('should handle case insensitive extensions', () => {
            assert.strictEqual(getFileType('test.JS'), 'javascript');
            assert.strictEqual(getFileType('test.PY'), 'python');
            assert.strictEqual(getFileType('test.MD'), 'markdown');
        });

        test('should return text for unknown extensions', () => {
            assert.strictEqual(getFileType('test.xyz'), 'text');
            assert.strictEqual(getFileType('test.unknown'), 'text');
        });

        test('should handle files without extensions', () => {
            assert.strictEqual(getFileType('README'), 'text');
            assert.strictEqual(getFileType('Makefile'), 'text');
        });
    });

    suite('formatBytes', () => {
        test('should format bytes correctly', () => {
            assert.strictEqual(formatBytes(0), '0 Bytes');
            assert.strictEqual(formatBytes(1024), '1 KB');
            assert.strictEqual(formatBytes(1048576), '1 MB');
            assert.strictEqual(formatBytes(1073741824), '1 GB');
        });

        test('should handle decimal places', () => {
            assert.strictEqual(formatBytes(1536), '1.5 KB'); // 1.5 * 1024
            assert.strictEqual(formatBytes(1572864), '1.5 MB'); // 1.5 * 1024 * 1024
        });

        test('should handle negative numbers', () => {
            assert.strictEqual(formatBytes(-1024), '-1 KB');
        });

        test('should handle very large numbers', () => {
            const result = formatBytes(1099511627776); // 1 TB
            assert.ok(result.includes('TB'));
        });
    });

    suite('debounce', () => {
        test('should delay function execution', (done) => {
            let callCount = 0;
            const debouncedFn = debounce(() => {
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
            let receivedArgs: any[] = [];
            const debouncedFn = debounce((...args: any[]) => {
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
            const debouncedFn = debounce(() => {
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

            assert.strictEqual(isLargeSelection(smallText), false);
            assert.strictEqual(isLargeSelection(mediumText), false);
            assert.strictEqual(isLargeSelection(largeText), true);
        });

        test('should handle empty strings', () => {
            assert.strictEqual(isLargeSelection(''), false);
        });

        test('should use custom threshold', () => {
            const text = 'a'.repeat(5000);
            assert.strictEqual(isLargeSelection(text, 1000), true);
            assert.strictEqual(isLargeSelection(text, 10000), false);
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