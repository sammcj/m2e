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
suite('Extension Integration Test Suite', () => {
    let mockServer;
    suiteSetup(async () => {
        // Start mock server for testing
        mockServer = new mockServer_1.MockM2EServer(18184);
        await mockServer.start();
        // Setup test workspace with mock server port
        await (0, testUtils_1.setupTestWorkspace)({
            serverPort: 18184,
            enableDiagnostics: true,
            debugLogging: true
        });
        // Wait for extension to activate
        await (0, testUtils_1.delay)(1000);
    });
    suiteTeardown(async () => {
        await (0, testUtils_1.cleanupTestWorkspace)();
        await mockServer.stop();
    });
    suite('Extension Activation', () => {
        test('should activate extension on command execution', async () => {
            // Extension should be activated
            const extension = vscode.extensions.getExtension('sammcj.m2e-vscode');
            assert.ok(extension);
            if (!extension.isActive) {
                await extension.activate();
            }
            assert.ok(extension.isActive);
        });
        test('should register all commands', async () => {
            const commands = await vscode.commands.getCommands();
            const expectedCommands = [
                'm2e.convertSelection',
                'm2e.convertFile',
                'm2e.convertCommentsOnly',
                'm2e.convertAndPreview',
                'm2e.restartServer',
                'm2e.ignoreWord',
                'm2e.manageIgnoreList',
                'm2e.refreshDiagnostics'
            ];
            for (const command of expectedCommands) {
                assert.ok(commands.includes(command), `Command ${command} not registered`);
            }
        });
    });
    suite('Convert Selection Command', () => {
        test('should convert selected American text to British', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color of the organization is great');
            const editor = await vscode.window.showTextDocument(document);
            // Select "color"
            editor.selection = new vscode.Selection(new vscode.Position(0, 4), new vscode.Position(0, 9));
            // Execute convert selection command
            await (0, testUtils_1.executeCommand)('m2e.convertSelection');
            // Wait for conversion to complete
            await (0, testUtils_1.delay)(500);
            // Check if text was converted
            const updatedText = document.getText();
            assert.ok(updatedText.includes('colour'), 'Text should contain "colour"');
        });
        test('should show warning when no text is selected', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            const editor = await vscode.window.showTextDocument(document);
            // Clear selection
            editor.selection = new vscode.Selection(new vscode.Position(0, 0), new vscode.Position(0, 0));
            // Mock showWarningMessage to capture the call
            let warningShown = false;
            const originalShowWarning = vscode.window.showWarningMessage;
            vscode.window.showWarningMessage = async (message) => {
                warningShown = true;
                assert.ok(message.includes('No text selected'));
                return undefined;
            };
            await (0, testUtils_1.executeCommand)('m2e.convertSelection');
            await (0, testUtils_1.delay)(100);
            assert.ok(warningShown, 'Warning should be shown for empty selection');
            // Restore original function
            vscode.window.showWarningMessage = originalShowWarning;
        });
        test('should handle large text selections', async () => {
            const largeText = 'The color is great. '.repeat(10000); // ~200KB
            const document = await (0, testUtils_1.createTestDocument)(largeText);
            const editor = await vscode.window.showTextDocument(document);
            // Select all text
            editor.selection = new vscode.Selection(new vscode.Position(0, 0), document.positionAt(largeText.length));
            // Mock showWarningMessage for large selection warning
            let warningShown = false;
            const originalShowWarning = vscode.window.showWarningMessage;
            vscode.window.showWarningMessage = async (message, ...items) => {
                if (message.includes('Large selection detected')) {
                    warningShown = true;
                    return 'Continue'; // Simulate user clicking Continue
                }
                return undefined;
            };
            await (0, testUtils_1.executeCommand)('m2e.convertSelection');
            await (0, testUtils_1.delay)(1000); // Allow more time for large text
            assert.ok(warningShown, 'Warning should be shown for large selection');
            // Restore original function
            vscode.window.showWarningMessage = originalShowWarning;
        });
    });
    suite('Convert File Command', () => {
        test('should convert entire file content', async () => {
            const originalText = 'The color of the organization center is great';
            const document = await (0, testUtils_1.createTestDocument)(originalText);
            await vscode.window.showTextDocument(document);
            // Execute convert file command
            await (0, testUtils_1.executeCommand)('m2e.convertFile');
            // Wait for conversion to complete
            await (0, testUtils_1.delay)(500);
            // Check if text was converted
            const updatedText = document.getText();
            assert.ok(updatedText.includes('colour'), 'Text should contain "colour"');
            assert.ok(updatedText.includes('organisation'), 'Text should contain "organisation"');
            assert.ok(updatedText.includes('centre'), 'Text should contain "centre"');
        });
        test('should show info message when no changes needed', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The colour is great'); // Already British
            await vscode.window.showTextDocument(document);
            // Mock showInformationMessage
            let infoShown = false;
            const originalShowInfo = vscode.window.showInformationMessage;
            vscode.window.showInformationMessage = async (message) => {
                if (message.includes('No changes needed')) {
                    infoShown = true;
                }
                return undefined;
            };
            await (0, testUtils_1.executeCommand)('m2e.convertFile');
            await (0, testUtils_1.delay)(500);
            assert.ok(infoShown, 'Info message should be shown when no changes needed');
            // Restore original function
            vscode.window.showInformationMessage = originalShowInfo;
        });
        test('should handle unsaved changes prompt', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            const editor = await vscode.window.showTextDocument(document);
            // Make a change to mark document as dirty
            await editor.edit(editBuilder => {
                editBuilder.insert(new vscode.Position(0, 0), ' ');
            });
            assert.ok(document.isDirty, 'Document should be marked as dirty');
            // Mock showWarningMessage for unsaved changes
            let warningShown = false;
            const originalShowWarning = vscode.window.showWarningMessage;
            vscode.window.showWarningMessage = async (message, ...items) => {
                if (message.includes('unsaved changes')) {
                    warningShown = true;
                    return 'Continue Without Saving';
                }
                return undefined;
            };
            await (0, testUtils_1.executeCommand)('m2e.convertFile');
            await (0, testUtils_1.delay)(500);
            assert.ok(warningShown, 'Warning should be shown for unsaved changes');
            // Restore original function
            vscode.window.showWarningMessage = originalShowWarning;
        });
    });
    suite('Convert Comments Only Command', () => {
        test('should convert only comments in JavaScript files', async () => {
            const originalText = `// The color is great
function test() {
    // Another color comment
    return "color in string";
}`;
            const document = await (0, testUtils_1.createTestDocument)(originalText, 'javascript');
            await vscode.window.showTextDocument(document);
            // Execute convert comments only command
            await (0, testUtils_1.executeCommand)('m2e.convertCommentsOnly');
            // Wait for conversion to complete
            await (0, testUtils_1.delay)(500);
            const updatedText = document.getText();
            assert.ok(updatedText.includes('// The colour is great'), 'Comment should be converted');
            assert.ok(updatedText.includes('// Another colour comment'), 'Second comment should be converted');
            assert.ok(updatedText.includes('"color in string"'), 'String should NOT be converted');
        });
        test('should show warning for non-code files', async () => {
            const document = await (0, testUtils_1.createTestDocument)('The color is great', 'plaintext');
            await vscode.window.showTextDocument(document);
            // Mock showWarningMessage
            let warningShown = false;
            const originalShowWarning = vscode.window.showWarningMessage;
            vscode.window.showWarningMessage = async (message) => {
                if (message.includes('recognised programming language')) {
                    warningShown = true;
                }
                return undefined;
            };
            await (0, testUtils_1.executeCommand)('m2e.convertCommentsOnly');
            await (0, testUtils_1.delay)(100);
            assert.ok(warningShown, 'Warning should be shown for non-code files');
            // Restore original function
            vscode.window.showWarningMessage = originalShowWarning;
        });
    });
    suite('Server Management', () => {
        test('should restart server successfully', async () => {
            let infoShown = false;
            const originalShowInfo = vscode.window.showInformationMessage;
            vscode.window.showInformationMessage = async (message) => {
                if (message.includes('Server restarted successfully')) {
                    infoShown = true;
                }
                return undefined;
            };
            await (0, testUtils_1.executeCommand)('m2e.restartServer');
            await (0, testUtils_1.delay)(1000); // Allow time for server restart
            assert.ok(infoShown, 'Success message should be shown after restart');
            // Restore original function
            vscode.window.showInformationMessage = originalShowInfo;
        });
    });
    suite('Error Handling', () => {
        test('should handle server connection errors gracefully', async () => {
            // Stop the mock server to simulate connection error
            await mockServer.stop();
            const document = await (0, testUtils_1.createTestDocument)('The color is great');
            const editor = await vscode.window.showTextDocument(document);
            // Select text
            editor.selection = new vscode.Selection(new vscode.Position(0, 4), new vscode.Position(0, 9));
            // Mock showErrorMessage
            let errorShown = false;
            const originalShowError = vscode.window.showErrorMessage;
            vscode.window.showErrorMessage = async (message) => {
                if (message.includes('Failed to')) {
                    errorShown = true;
                }
                return undefined;
            };
            await (0, testUtils_1.executeCommand)('m2e.convertSelection');
            await (0, testUtils_1.delay)(500);
            assert.ok(errorShown, 'Error message should be shown when server is unavailable');
            // Restore original function and restart server
            vscode.window.showErrorMessage = originalShowError;
            await mockServer.start();
        });
        test('should handle empty files gracefully', async () => {
            const document = await (0, testUtils_1.createTestDocument)('');
            await vscode.window.showTextDocument(document);
            // Mock showWarningMessage
            let warningShown = false;
            const originalShowWarning = vscode.window.showWarningMessage;
            vscode.window.showWarningMessage = async (message) => {
                if (message.includes('empty')) {
                    warningShown = true;
                }
                return undefined;
            };
            await (0, testUtils_1.executeCommand)('m2e.convertFile');
            await (0, testUtils_1.delay)(100);
            assert.ok(warningShown, 'Warning should be shown for empty files');
            // Restore original function
            vscode.window.showWarningMessage = originalShowWarning;
        });
    });
    suite('Configuration Changes', () => {
        test('should respond to configuration changes', async () => {
            const config = vscode.workspace.getConfiguration('m2e');
            // Change a setting
            await config.update('enableDiagnostics', false, vscode.ConfigurationTarget.Workspace);
            await (0, testUtils_1.delay)(100);
            // Verify setting was changed
            const updatedConfig = vscode.workspace.getConfiguration('m2e');
            assert.strictEqual(updatedConfig.get('enableDiagnostics'), false);
            // Reset setting
            await config.update('enableDiagnostics', true, vscode.ConfigurationTarget.Workspace);
        });
    });
});
//# sourceMappingURL=extension.test.js.map