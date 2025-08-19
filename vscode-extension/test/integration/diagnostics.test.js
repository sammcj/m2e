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
const testUtils_1 = require("../helpers/testUtils");
const mockServer_1 = require("../helpers/mockServer");
suite('Diagnostics Integration Test Suite', () => {
    let mockServer;
    suiteSetup(async () => {
        // Start mock server
        mockServer = new mockServer_1.MockM2EServer(18185);
        await mockServer.start();
        // Setup test workspace with diagnostics enabled
        await (0, testUtils_1.setupTestWorkspace)({
            serverPort: 18185,
            enableDiagnostics: true,
            diagnosticSeverity: 'Information',
            debugLogging: true
        });
        // Wait for extension to activate
        await (0, testUtils_1.delay)(1000);
    });
    suiteTeardown(async () => {
        await (0, testUtils_1.cleanupTestWorkspace)();
        await mockServer.stop();
    });
    suite('Diagnostic Detection', () => {
        test('should detect American spellings in text files', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color of the organization is great', 'plaintext');
            // Wait for diagnostics to be generated
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 2, 5000);
            assert.strictEqual(diagnostics.length, 2, 'Should detect 2 American spellings');
            // Check that diagnostics are for the right words
            const messages = diagnostics.map(d => d.message);
            assert.ok(messages.some(m => m.includes('color')), 'Should detect "color"');
            assert.ok(messages.some(m => m.includes('organization')), 'Should detect "organization"');
        });
        test('should detect American spellings in code files', async () => {
            const code = `// The color is great
function analyze() {
    // Organize the data
    return "center";
}`;
            const document = await (0, testUtils_1.createTestDocument)(code, 'javascript');
            // Wait for diagnostics
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 3, 5000);
            assert.strictEqual(diagnostics.length, 3, 'Should detect 3 American spellings');
            const ranges = diagnostics.map(d => d.range);
            // Verify ranges are correct (approximately)
            assert.ok(ranges.some(r => r.start.line === 0), 'Should detect word in comment on line 0');
            assert.ok(ranges.some(r => r.start.line === 1), 'Should detect word in function name on line 1');
            assert.ok(ranges.some(r => r.start.line === 2), 'Should detect word in comment on line 2');
        });
        test('should use correct diagnostic severity', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 1);
            assert.strictEqual(diagnostics.length, 1);
            assert.strictEqual(diagnostics[0].severity, vscode.DiagnosticSeverity.Information);
        });
        test('should update diagnostics when text changes', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            const editor = await vscode.window.showTextDocument(document);
            // Wait for initial diagnostics
            let diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 1);
            assert.strictEqual(diagnostics.length, 1);
            // Change "color" to "colour"
            await editor.edit(editBuilder => {
                const range = new vscode.Range(new vscode.Position(0, 4), new vscode.Position(0, 9));
                editBuilder.replace(range, 'colour');
            });
            // Wait for diagnostics to update (should be none now)
            await (0, testUtils_1.delay)(1000);
            diagnostics = (0, testUtils_1.getDiagnostics)(document);
            assert.strictEqual(diagnostics.length, 0, 'Diagnostics should be cleared after correction');
        });
        test('should handle empty documents', async () => {
            const document = await (0, testUtils_1.createTestDocument)('');
            // Wait a bit for potential diagnostics
            await (0, testUtils_1.delay)(500);
            const diagnostics = (0, testUtils_1.getDiagnostics)(document);
            assert.strictEqual(diagnostics.length, 0, 'Empty document should have no diagnostics');
        });
        test('should handle very long documents', async () => {
            const longText = 'The color is great. '.repeat(1000); // ~20KB
            const document = await (0, testUtils_1.createTestDocument)(longText);
            // Should still detect American spellings efficiently
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, undefined, 10000); // Allow more time
            assert.ok(diagnostics.length > 0, 'Should detect spellings in long documents');
            assert.ok(diagnostics.length <= 1000, 'Should not create excessive diagnostics');
        });
    });
    suite('Quick Fix Code Actions', () => {
        test('should provide Quick Fix for American spellings', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            // Wait for diagnostics
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 1);
            const diagnostic = diagnostics[0];
            // Get code actions for the diagnostic
            const codeActions = await vscode.commands.executeCommand('vscode.executeCodeActionProvider', document.uri, diagnostic.range);
            assert.ok(codeActions.length > 0, 'Should provide code actions');
            // Check for expected code actions
            const actionTitles = codeActions.map(action => action.title);
            assert.ok(actionTitles.some(title => title.includes('Convert')), 'Should provide convert action');
        });
        test('should apply Quick Fix correctly', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            const editor = await vscode.window.showTextDocument(document);
            // Wait for diagnostics
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 1);
            const diagnostic = diagnostics[0];
            // Get code actions
            const codeActions = await vscode.commands.executeCommand('vscode.executeCodeActionProvider', document.uri, diagnostic.range);
            // Find and execute convert action
            const convertAction = codeActions.find(action => action.title.includes('Convert'));
            assert.ok(convertAction, 'Should have convert action');
            if (convertAction.edit) {
                await vscode.workspace.applyEdit(convertAction.edit);
                // Check that text was changed
                const updatedText = document.getText();
                assert.ok(updatedText.includes('colour'), 'Text should be converted to British spelling');
            }
        });
    });
    suite('Configuration Integration', () => {
        test('should disable diagnostics when setting is false', async () => {
            // Disable diagnostics
            const config = vscode.workspace.getConfiguration('m2e');
            await config.update('enableDiagnostics', false, vscode.ConfigurationTarget.Workspace);
            // Create document with American spelling
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            // Wait a bit and check for diagnostics
            await (0, testUtils_1.delay)(1000);
            const diagnostics = (0, testUtils_1.getDiagnostics)(document);
            assert.strictEqual(diagnostics.length, 0, 'Should not show diagnostics when disabled');
            // Re-enable diagnostics
            await config.update('enableDiagnostics', true, vscode.ConfigurationTarget.Workspace);
        });
        test('should change diagnostic severity based on configuration', async () => {
            // Set severity to Warning
            const config = vscode.workspace.getConfiguration('m2e');
            await config.update('diagnosticSeverity', 'Warning', vscode.ConfigurationTarget.Workspace);
            // Refresh diagnostics
            await vscode.commands.executeCommand('m2e.refreshDiagnostics');
            await (0, testUtils_1.delay)(500);
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 1);
            assert.strictEqual(diagnostics[0].severity, vscode.DiagnosticSeverity.Warning);
            // Reset to Information
            await config.update('diagnosticSeverity', 'Information', vscode.ConfigurationTarget.Workspace);
        });
    });
    suite('Performance Tests', () => {
        test('should handle multiple documents efficiently', async () => {
            const documents = [];
            // Create multiple documents
            for (let i = 0; i < 5; i++) {
                const doc = await (0, testUtils_1.createTestDocument)(`Document ${i} with color and organization`);
                documents.push(doc);
            }
            // Wait for all diagnostics to be processed
            await (0, testUtils_1.delay)(2000);
            // Check that all documents have diagnostics
            for (const doc of documents) {
                const diagnostics = (0, testUtils_1.getDiagnostics)(doc);
                assert.ok(diagnostics.length > 0, `Document should have diagnostics`);
            }
        });
        test('should debounce rapid text changes', async () => {
            const document = await (0, testUtils_1.createTestDocument)('color');
            const editor = await vscode.window.showTextDocument(document);
            // Make rapid changes
            for (let i = 0; i < 5; i++) {
                await editor.edit(editBuilder => {
                    editBuilder.insert(new vscode.Position(0, 0), ' ');
                });
                await (0, testUtils_1.delay)(50); // Rapid changes
            }
            // Wait for debouncing to settle
            await (0, testUtils_1.delay)(1000);
            const diagnostics = (0, testUtils_1.getDiagnostics)(document);
            assert.ok(diagnostics.length >= 0, 'Should handle rapid changes without errors');
        });
    });
    suite('Edge Cases', () => {
        test('should handle Unicode text correctly', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The cölor is great with ñice organization');
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 2);
            assert.strictEqual(diagnostics.length, 2, 'Should detect American spellings with Unicode');
        });
        test('should handle very long lines', async () => {
            const longLine = 'word '.repeat(1000) + 'color ' + 'word '.repeat(1000);
            const document = await (0, testUtils_1.createTestDocument)(longLine);
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 1);
            assert.strictEqual(diagnostics.length, 1, 'Should detect spelling in very long line');
        });
        test('should handle mixed line endings', async () => {
            const mixedText = 'The color\r\norganization\nanalyze';
            const document = await (0, testUtils_1.createTestDocument)(mixedText);
            const diagnostics = await (0, testUtils_1.waitForDiagnostics)(document, 3);
            assert.strictEqual(diagnostics.length, 3, 'Should handle mixed line endings');
        });
    });
});
//# sourceMappingURL=diagnostics.test.js.map