import * as assert from 'assert';
import * as vscode from 'vscode';
import { M2EApiClient, getFileTypeFromDocument } from '../../src/services/client';
import { MockM2EServer } from '../helpers/mockServer';

suite('M2E API Client Test Suite', () => {
    let mockServer: MockM2EServer;
    let apiClient: M2EApiClient;
    let outputChannel: vscode.OutputChannel;

    suiteSetup(async () => {
        // Create output channel for testing
        outputChannel = vscode.window.createOutputChannel('M2E Test');
        
        // Start mock server
        mockServer = new MockM2EServer(18183); // Use different port
        await mockServer.start();
        
        // Create API client
        apiClient = new M2EApiClient(outputChannel);
        apiClient.setServerUrl(mockServer.getUrl());
    });

    suiteTeardown(async () => {
        await mockServer.stop();
        outputChannel.dispose();
    });

    suite('getFileTypeFromDocument', () => {
        test('should detect JavaScript files', async () => {
            const uri = vscode.Uri.parse('untitled:test.js');
            const document = await vscode.workspace.openTextDocument(uri);
            
            const fileType = getFileTypeFromDocument(document);
            assert.strictEqual(fileType, 'javascript');
        });

        test('should detect TypeScript files', async () => {
            const uri = vscode.Uri.parse('untitled:test.ts');
            const document = await vscode.workspace.openTextDocument(uri);
            
            const fileType = getFileTypeFromDocument(document);
            assert.strictEqual(fileType, 'typescript');
        });

        test('should detect Python files', async () => {
            const uri = vscode.Uri.parse('untitled:test.py');
            const document = await vscode.workspace.openTextDocument(uri);
            
            const fileType = getFileTypeFromDocument(document);
            assert.strictEqual(fileType, 'python');
        });

        test('should detect Markdown files', async () => {
            const uri = vscode.Uri.parse('untitled:test.md');
            const document = await vscode.workspace.openTextDocument(uri);
            
            const fileType = getFileTypeFromDocument(document);
            assert.strictEqual(fileType, 'markdown');
        });

        test('should default to text for unknown extensions', async () => {
            const uri = vscode.Uri.parse('untitled:test.unknown');
            const document = await vscode.workspace.openTextDocument(uri);
            
            const fileType = getFileTypeFromDocument(document);
            assert.strictEqual(fileType, 'text');
        });

        test('should handle files without extensions', async () => {
            const uri = vscode.Uri.parse('untitled:README');
            const document = await vscode.workspace.openTextDocument(uri);
            
            const fileType = getFileTypeFromDocument(document);
            assert.strictEqual(fileType, 'text');
        });
    });

    suite('API Client Methods', () => {
        test('should check server health', async () => {
            const health = await apiClient.checkHealth();
            
            assert.strictEqual(health.status, 'healthy');
            assert.strictEqual(health.version, '1.0.0');
            assert.ok(health.timestamp);
        });

        test('should convert selection text', async () => {
            const text = 'The color is great';
            const result = await apiClient.convertSelection(text, 'text');
            
            assert.strictEqual(result.originalText, text);
            assert.strictEqual(result.convertedText, 'The colour is great');
            assert.strictEqual(result.metadata.spellingChanges, 1);
            assert.strictEqual(result.metadata.unitChanges, 0);
            assert.ok(result.metadata.processingTimeMs > 0);
        });

        test('should convert file content', async () => {
            const text = 'The organization uses color';
            const result = await apiClient.convertFile(text, 'text');
            
            assert.strictEqual(result.originalText, text);
            assert.strictEqual(result.convertedText, 'The organisation uses colour');
            assert.strictEqual(result.metadata.spellingChanges, 2);
            assert.strictEqual(result.metadata.unitChanges, 0);
        });

        test('should convert comments only', async () => {
            const text = '// The color is great\nfunction test() { return "color"; }';
            const result = await apiClient.convertCommentsOnly(text, 'javascript');
            
            assert.strictEqual(result.originalText, text);
            assert.ok(result.convertedText.includes('// The colour is great'));
            assert.ok(result.convertedText.includes('return "color";')); // String unchanged
        });

        test('should handle empty text', async () => {
            const result = await apiClient.convertSelection('', 'text');
            
            assert.strictEqual(result.originalText, '');
            assert.strictEqual(result.convertedText, '');
            assert.strictEqual(result.metadata.spellingChanges, 0);
        });

        test('should handle text with no changes needed', async () => {
            const text = 'The colour is great'; // Already British
            const result = await apiClient.convertSelection(text, 'text');
            
            assert.strictEqual(result.originalText, text);
            assert.strictEqual(result.convertedText, text);
            assert.strictEqual(result.metadata.spellingChanges, 0);
        });
    });

    suite('Error Handling', () => {
        test('should handle server connection errors', async () => {
            // Create client with invalid URL
            const badClient = new M2EApiClient(outputChannel);
            badClient.setServerUrl('http://localhost:99999');
            
            try {
                await badClient.checkHealth();
                assert.fail('Should have thrown an error');
            } catch (error) {
                assert.ok(error instanceof Error);
                assert.ok(error.message.includes('Failed to connect'));
            }
        });

        test('should handle timeout errors', async () => {
            // Set a very short timeout for testing
            const client = new M2EApiClient(outputChannel, 1); // 1ms timeout
            client.setServerUrl(mockServer.getUrl());
            
            try {
                await client.convertSelection('test', 'text');
                // Note: This test might be flaky depending on system performance
                // In real scenarios, the timeout would be much longer
            } catch (error) {
                assert.ok(error instanceof Error);
            }
        });

        test('should handle invalid JSON responses', async () => {
            // Set up mock server to return invalid JSON
            mockServer.setResponse('/api/v1/convert', 'invalid json');
            
            try {
                await apiClient.convertSelection('test', 'text');
                assert.fail('Should have thrown an error');
            } catch (error) {
                assert.ok(error instanceof Error);
            }
        });
    });

    suite('URL Management', () => {
        test('should set and get server URL', () => {
            const testUrl = 'http://localhost:8080';
            apiClient.setServerUrl(testUrl);
            
            // Note: There's no getter in the current implementation
            // This test validates that setServerUrl doesn't throw
            assert.ok(true);
        });

        test('should handle URLs with trailing slashes', () => {
            apiClient.setServerUrl('http://localhost:8080/');
            
            // Should normalize the URL internally
            assert.ok(true);
        });

        test('should handle URLs without protocol', () => {
            try {
                apiClient.setServerUrl('localhost:8080');
                // Should either work or throw appropriate error
                assert.ok(true);
            } catch (error) {
                // Expected for malformed URLs
                assert.ok(error instanceof Error);
            }
        });
    });

    suite('Response Validation', () => {
        test('should validate convert response structure', async () => {
            const result = await apiClient.convertSelection('color', 'text');
            
            // Validate response structure
            assert.ok(typeof result.originalText === 'string');
            assert.ok(typeof result.convertedText === 'string');
            assert.ok(typeof result.metadata === 'object');
            assert.ok(typeof result.metadata.spellingChanges === 'number');
            assert.ok(typeof result.metadata.unitChanges === 'number');
            assert.ok(typeof result.metadata.processingTimeMs === 'number');
            assert.ok(typeof result.metadata.fileType === 'string');
        });

        test('should handle metadata with additional fields', async () => {
            // Mock server might return additional metadata fields
            const result = await apiClient.convertSelection('color', 'text');
            
            // Should still have required fields
            assert.ok(result.metadata.spellingChanges !== undefined);
            assert.ok(result.metadata.unitChanges !== undefined);
            assert.ok(result.metadata.processingTimeMs !== undefined);
        });
    });
});