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
exports.MockConsole = void 0;
exports.createTestDocument = createTestDocument;
exports.delay = delay;
exports.waitFor = waitFor;
exports.setupTestWorkspace = setupTestWorkspace;
exports.cleanupTestWorkspace = cleanupTestWorkspace;
exports.createSelection = createSelection;
exports.getDocumentText = getDocumentText;
exports.assertTextEqual = assertTextEqual;
exports.createTestFile = createTestFile;
exports.deleteTestFile = deleteTestFile;
exports.getDiagnostics = getDiagnostics;
exports.waitForDiagnostics = waitForDiagnostics;
exports.executeCommand = executeCommand;
const vscode = __importStar(require("vscode"));
const path = __importStar(require("path"));
const fs = __importStar(require("fs"));
/**
 * Utility functions for VSCode extension testing
 */
/**
 * Create a temporary test document
 */
async function createTestDocument(content, language = 'plaintext') {
    const uri = vscode.Uri.parse(`untitled:test.${getFileExtension(language)}`);
    const document = await vscode.workspace.openTextDocument(uri);
    // Insert content into the document
    const editor = await vscode.window.showTextDocument(document);
    await editor.edit(editBuilder => {
        editBuilder.insert(new vscode.Position(0, 0), content);
    });
    return document;
}
/**
 * Get file extension for language
 */
function getFileExtension(language) {
    const extensions = {
        'javascript': 'js',
        'typescript': 'ts',
        'python': 'py',
        'go': 'go',
        'java': 'java',
        'markdown': 'md',
        'plaintext': 'txt'
    };
    return extensions[language] || 'txt';
}
/**
 * Wait for a specified amount of time
 */
function delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
/**
 * Wait for a condition to be true
 */
async function waitFor(condition, timeout = 5000, interval = 100) {
    const start = Date.now();
    while (Date.now() - start < timeout) {
        if (await condition()) {
            return;
        }
        await delay(interval);
    }
    throw new Error(`Condition not met within ${timeout}ms`);
}
/**
 * Setup test workspace with configuration
 */
async function setupTestWorkspace(config = {}) {
    const configuration = vscode.workspace.getConfiguration('m2e');
    // Set test configuration
    const defaultConfig = {
        enableDiagnostics: true,
        diagnosticSeverity: 'Information',
        enableUnitConversion: true,
        serverPort: 18182, // Use different port for testing
        showStatusBar: false, // Disable for testing
        debugLogging: true
    };
    const testConfig = { ...defaultConfig, ...config };
    for (const [key, value] of Object.entries(testConfig)) {
        await configuration.update(key, value, vscode.ConfigurationTarget.Workspace);
    }
}
/**
 * Clean up test workspace
 */
async function cleanupTestWorkspace() {
    // Close all open editors
    await vscode.commands.executeCommand('workbench.action.closeAllEditors');
    // Reset configuration
    const configuration = vscode.workspace.getConfiguration('m2e');
    const keys = [
        'enableDiagnostics',
        'diagnosticSeverity',
        'enableUnitConversion',
        'serverPort',
        'showStatusBar',
        'debugLogging'
    ];
    for (const key of keys) {
        await configuration.update(key, undefined, vscode.ConfigurationTarget.Workspace);
    }
}
/**
 * Create test selection in editor
 */
async function createSelection(editor, startLine, startChar, endLine, endChar) {
    const start = new vscode.Position(startLine, startChar);
    const end = new vscode.Position(endLine, endChar);
    const selection = new vscode.Selection(start, end);
    editor.selection = selection;
    return selection;
}
/**
 * Get text from document
 */
function getDocumentText(document) {
    return document.getText();
}
/**
 * Assert that two texts are equal (with better error messages)
 */
function assertTextEqual(actual, expected, message) {
    if (actual !== expected) {
        const baseMessage = message || 'Text mismatch';
        const details = `\nExpected: "${expected}"\nActual:   "${actual}"`;
        throw new Error(baseMessage + details);
    }
}
/**
 * Create a test file in the workspace
 */
async function createTestFile(relativePath, content) {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
        throw new Error('No workspace folder available');
    }
    const filePath = path.join(workspaceFolder.uri.fsPath, relativePath);
    const dir = path.dirname(filePath);
    // Ensure directory exists
    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
    }
    // Write file
    fs.writeFileSync(filePath, content);
    return vscode.Uri.file(filePath);
}
/**
 * Delete test file
 */
async function deleteTestFile(uri) {
    if (fs.existsSync(uri.fsPath)) {
        fs.unlinkSync(uri.fsPath);
    }
}
/**
 * Get diagnostics for a document
 */
function getDiagnostics(document) {
    return vscode.languages.getDiagnostics(document.uri);
}
/**
 * Wait for diagnostics to be updated
 */
async function waitForDiagnostics(document, expectedCount, timeout = 3000) {
    await waitFor(() => {
        const diagnostics = getDiagnostics(document);
        return expectedCount !== undefined ? diagnostics.length === expectedCount : diagnostics.length > 0;
    }, timeout);
    return getDiagnostics(document);
}
/**
 * Execute command and wait for completion
 */
async function executeCommand(command, ...args) {
    return await vscode.commands.executeCommand(command, ...args);
}
/**
 * Mock console for capturing logs during tests
 */
class MockConsole {
    constructor() {
        this.logs = [];
        this.errors = [];
        this.originalLog = console.log;
        this.originalError = console.error;
    }
    start() {
        console.log = (...args) => {
            this.logs.push(args.join(' '));
        };
        console.error = (...args) => {
            this.errors.push(args.join(' '));
        };
    }
    stop() {
        console.log = this.originalLog;
        console.error = this.originalError;
    }
    clear() {
        this.logs = [];
        this.errors = [];
    }
    hasLog(message) {
        return this.logs.some(log => log.includes(message));
    }
    hasError(message) {
        return this.errors.some(error => error.includes(message));
    }
}
exports.MockConsole = MockConsole;
//# sourceMappingURL=testUtils.js.map