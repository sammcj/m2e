import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';

/**
 * Utility functions for VSCode extension testing
 */

/**
 * Create a temporary test document
 */
export async function createTestDocument(content: string, language: string = 'plaintext'): Promise<vscode.TextDocument> {
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
function getFileExtension(language: string): string {
    const extensions: { [key: string]: string } = {
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
export function delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Wait for a condition to be true
 */
export async function waitFor(
    condition: () => boolean | Promise<boolean>,
    timeout: number = 5000,
    interval: number = 100
): Promise<void> {
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
export async function setupTestWorkspace(config: any = {}): Promise<void> {
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
export async function cleanupTestWorkspace(): Promise<void> {
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
export async function createSelection(
    editor: vscode.TextEditor,
    startLine: number,
    startChar: number,
    endLine: number,
    endChar: number
): Promise<vscode.Selection> {
    const start = new vscode.Position(startLine, startChar);
    const end = new vscode.Position(endLine, endChar);
    const selection = new vscode.Selection(start, end);
    
    editor.selection = selection;
    return selection;
}

/**
 * Get text from document
 */
export function getDocumentText(document: vscode.TextDocument): string {
    return document.getText();
}

/**
 * Assert that two texts are equal (with better error messages)
 */
export function assertTextEqual(actual: string, expected: string, message?: string): void {
    if (actual !== expected) {
        const baseMessage = message || 'Text mismatch';
        const details = `\nExpected: "${expected}"\nActual:   "${actual}"`;
        throw new Error(baseMessage + details);
    }
}

/**
 * Create a test file in the workspace
 */
export async function createTestFile(
    relativePath: string,
    content: string
): Promise<vscode.Uri> {
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
export async function deleteTestFile(uri: vscode.Uri): Promise<void> {
    if (fs.existsSync(uri.fsPath)) {
        fs.unlinkSync(uri.fsPath);
    }
}

/**
 * Get diagnostics for a document
 */
export function getDiagnostics(document: vscode.TextDocument): vscode.Diagnostic[] {
    return vscode.languages.getDiagnostics(document.uri);
}

/**
 * Wait for diagnostics to be updated
 */
export async function waitForDiagnostics(
    document: vscode.TextDocument,
    expectedCount?: number,
    timeout: number = 3000
): Promise<vscode.Diagnostic[]> {
    await waitFor(
        () => {
            const diagnostics = getDiagnostics(document);
            return expectedCount !== undefined ? diagnostics.length === expectedCount : diagnostics.length > 0;
        },
        timeout
    );
    
    return getDiagnostics(document);
}

/**
 * Execute command and wait for completion
 */
export async function executeCommand(command: string, ...args: any[]): Promise<any> {
    return await vscode.commands.executeCommand(command, ...args);
}

/**
 * Mock console for capturing logs during tests
 */
export class MockConsole {
    public logs: string[] = [];
    public errors: string[] = [];
    
    private originalLog = console.log;
    private originalError = console.error;
    
    start(): void {
        console.log = (...args) => {
            this.logs.push(args.join(' '));
        };
        
        console.error = (...args) => {
            this.errors.push(args.join(' '));
        };
    }
    
    stop(): void {
        console.log = this.originalLog;
        console.error = this.originalError;
    }
    
    clear(): void {
        this.logs = [];
        this.errors = [];
    }
    
    hasLog(message: string): boolean {
        return this.logs.some(log => log.includes(message));
    }
    
    hasError(message: string): boolean {
        return this.errors.some(error => error.includes(message));
    }
}