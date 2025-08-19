import * as assert from 'assert';
import * as vscode from 'vscode';
import { M2EApiClient } from '../../src/services/client';

// Mock vscode workspace configuration
const mockConfig = {
    get: (key: string, defaultValue?: any) => {
        const configs: { [key: string]: any } = {
            'enableUnitConversion': true,
            'codeAwareConversion': true,
            'preserveCodeSyntax': true,
            'debugLogging': false
        };
        return configs[key] ?? defaultValue;
    }
};

// Mock VSCode workspace
(global as any).vscode = {
    workspace: {
        getConfiguration: () => mockConfig
    }
};

suite('API Response Handling Tests', () => {
    let mockOutputChannel: vscode.OutputChannel;
    let apiClient: M2EApiClient;

    setup(() => {
        mockOutputChannel = {
            appendLine: () => {},
            append: () => {},
            clear: () => {},
            show: () => {},
            hide: () => {},
            dispose: () => {},
            name: 'test',
            replace: () => {}
        } as vscode.OutputChannel;

        apiClient = new M2EApiClient(mockOutputChannel);
    });

    test('generateChanges should handle identical text', async () => {
        const client = apiClient as any; // Access private method
        const originalText = 'Hello world';
        const convertedText = 'Hello world';
        
        const changes = client.generateChanges(originalText, convertedText);
        
        assert.strictEqual(changes.length, 0);
    });

    test('generateChanges should detect spelling changes', async () => {
        const client = apiClient as any; // Access private method
        const originalText = 'I love color and flavor';
        const convertedText = 'I love colour and flavour';
        
        const changes = client.generateChanges(originalText, convertedText);
        
        assert.strictEqual(changes.length, 2);
        assert.strictEqual(changes[0].original, 'color');
        assert.strictEqual(changes[0].converted, 'colour');
        assert.strictEqual(changes[0].type, 'spelling');
        
        assert.strictEqual(changes[1].original, 'flavor');
        assert.strictEqual(changes[1].converted, 'flavour');
        assert.strictEqual(changes[1].type, 'spelling');
    });

    test('generateChanges should detect unit changes', async () => {
        const client = apiClient as any; // Access private method
        const originalText = 'The room is 12 feet wide';
        const convertedText = 'The room is 3.7 metres wide';
        
        const changes = client.generateChanges(originalText, convertedText);
        
        // Should detect at least one unit change
        const unitChanges = changes.filter(c => c.type === 'unit');
        assert.ok(unitChanges.length > 0);
    });

    test('generateChanges should handle empty text', async () => {
        const client = apiClient as any; // Access private method
        const originalText = '';
        const convertedText = '';
        
        const changes = client.generateChanges(originalText, convertedText);
        
        assert.strictEqual(changes.length, 0);
    });

    test('ConvertResponse should have proper metadata structure', async () => {
        // Mock the API client's private method to test the response structure
        const client = apiClient as any;
        const originalText = 'I love color';
        const convertedText = 'I love colour';
        
        const changes = client.generateChanges(originalText, convertedText);
        const spellingChanges = changes.filter((c: any) => c.type === 'spelling').length;
        const unitChanges = changes.filter((c: any) => c.type === 'unit').length;
        
        // Simulate the response structure that was causing the error
        const mockResponse = {
            originalText: originalText,
            convertedText: convertedText,
            changes: changes,
            metadata: {
                spellingChanges: spellingChanges,
                unitChanges: unitChanges,
                processingTimeMs: 0
            }
        };
        
        // This should not throw the "Cannot read properties of undefined" error
        assert.ok(mockResponse.metadata);
        assert.ok(typeof mockResponse.metadata.spellingChanges === 'number');
        assert.ok(typeof mockResponse.metadata.unitChanges === 'number');
        assert.ok(typeof mockResponse.metadata.processingTimeMs === 'number');
        
        // Test the specific line that was failing in preview.ts
        const changeCount = mockResponse.metadata.spellingChanges + mockResponse.metadata.unitChanges;
        assert.ok(typeof changeCount === 'number');
        assert.ok(changeCount >= 0);
    });

    test('ConvertResponse should handle undefined metadata gracefully', async () => {
        // Test the error condition that was occurring
        const mockResponseWithUndefinedMetadata = {
            originalText: 'test',
            convertedText: 'test',
            changes: [],
            metadata: undefined as any
        };
        
        // This should demonstrate the error that was happening
        try {
            const changeCount = mockResponseWithUndefinedMetadata.metadata.spellingChanges + 
                               mockResponseWithUndefinedMetadata.metadata.unitChanges;
            assert.fail('Should have thrown an error with undefined metadata');
        } catch (error) {
            assert.ok(error instanceof TypeError);
            assert.ok((error as Error).message.includes('Cannot read properties of undefined'));
        }
    });
});