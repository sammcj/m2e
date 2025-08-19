import * as vscode from 'vscode';
import { M2EApiClient, ConvertResponse, getFileTypeFromDocument } from '../services/client';

/**
 * Diagnostic provider for highlighting American spellings in documents
 */
export class M2EDiagnosticProvider {
    private diagnosticCollection: vscode.DiagnosticCollection;
    private apiClient: M2EApiClient;
    private outputChannel: vscode.OutputChannel;
    private ensureServerRunning: () => Promise<void>;
    
    // Debouncing and caching
    private updateTimers = new Map<string, ReturnType<typeof setTimeout>>();
    private lastProcessedVersion = new Map<string, number>();
    private readonly debounceDelay = 500; // ms
    
    // Workspace ignore list
    private workspaceIgnoreList = new Set<string>();
    private readonly ignoreListFile = '.m2e-ignore.json';

    constructor(
        context: vscode.ExtensionContext,
        apiClient: M2EApiClient,
        outputChannel: vscode.OutputChannel,
        ensureServerRunning: () => Promise<void>
    ) {
        this.apiClient = apiClient;
        this.outputChannel = outputChannel;
        this.ensureServerRunning = ensureServerRunning;
        
        // Create diagnostic collection
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection('m2e');
        context.subscriptions.push(this.diagnosticCollection);
        
        // Load workspace ignore list
        this.loadWorkspaceIgnoreList();
        
        // Set up event listeners
        this.setupEventListeners(context);
        
        // Process all currently open documents
        this.processOpenDocuments();
    }

    /**
     * Set up event listeners for document changes
     */
    private setupEventListeners(context: vscode.ExtensionContext): void {
        // Document opened
        context.subscriptions.push(
            vscode.workspace.onDidOpenTextDocument(document => {
                this.scheduleDocumentUpdate(document);
            })
        );

        // Document changed
        context.subscriptions.push(
            vscode.workspace.onDidChangeTextDocument(event => {
                this.scheduleDocumentUpdate(event.document);
            })
        );

        // Document closed
        context.subscriptions.push(
            vscode.workspace.onDidCloseTextDocument(document => {
                this.clearDocumentDiagnostics(document);
            })
        );

        // Configuration changed
        context.subscriptions.push(
            vscode.workspace.onDidChangeConfiguration(event => {
                if (event.affectsConfiguration('m2e')) {
                    this.processOpenDocuments();
                }
            })
        );

        // Workspace folders changed
        context.subscriptions.push(
            vscode.workspace.onDidChangeWorkspaceFolders(() => {
                this.loadWorkspaceIgnoreList();
            })
        );
    }

    /**
     * Schedule document update with debouncing
     */
    private scheduleDocumentUpdate(document: vscode.TextDocument): void {
        if (!this.shouldProcessDocument(document)) {
            return;
        }

        const uri = document.uri.toString();
        
        // Clear existing timer
        const existingTimer = this.updateTimers.get(uri);
        if (existingTimer) {
            clearTimeout(existingTimer);
        }

        // Set new timer
        const timer = setTimeout(() => {
            this.processDocument(document);
            this.updateTimers.delete(uri);
        }, this.debounceDelay);

        this.updateTimers.set(uri, timer);
    }

    /**
     * Check if document should be processed
     */
    private shouldProcessDocument(document: vscode.TextDocument): boolean {
        const config = vscode.workspace.getConfiguration('m2e');
        
        // Check if diagnostics are enabled
        if (!config.get<boolean>('enableDiagnostics', true)) {
            return false;
        }

        // Skip if document is too large (performance)
        const maxSize = 100000; // 100KB limit for real-time processing
        if (document.getText().length > maxSize) {
            return false;
        }

        // Check exclude patterns
        const excludePatterns = config.get<string[]>('excludePatterns', []);
        const relativePath = vscode.workspace.asRelativePath(document.uri);
        
        for (const pattern of excludePatterns) {
            if (this.matchesGlobPattern(relativePath, pattern)) {
                return false;
            }
        }

        // Check if document version changed
        const uri = document.uri.toString();
        const lastVersion = this.lastProcessedVersion.get(uri);
        if (lastVersion === document.version) {
            return false;
        }

        return true;
    }

    /**
     * Simple glob pattern matching
     */
    private matchesGlobPattern(path: string, pattern: string): boolean {
        // Convert glob pattern to regex
        const regexPattern = pattern
            .replace(/\*\*/g, '.*')
            .replace(/\*/g, '[^/]*')
            .replace(/\?/g, '[^/]');
        
        const regex = new RegExp(`^${regexPattern}$`);
        return regex.test(path);
    }

    /**
     * Process all currently open documents
     */
    private processOpenDocuments(): void {
        vscode.workspace.textDocuments.forEach(document => {
            this.scheduleDocumentUpdate(document);
        });
    }

    /**
     * Process a single document for American spellings
     */
    private async processDocument(document: vscode.TextDocument): Promise<void> {
        const uri = document.uri.toString();
        
        try {
            // Mark as processed
            this.lastProcessedVersion.set(uri, document.version);
            
            const text = document.getText();
            if (text.trim().length === 0) {
                this.diagnosticCollection.set(document.uri, []);
                return;
            }

            // Get file type for code-aware processing
            const fileType = getFileTypeFromDocument(document);
            
            // Ensure server is running
            await this.ensureServerRunning();
            
            // Get conversion results (this will identify American spellings)
            const response = await this.apiClient.convert({
                text,
                options: {
                    ...(fileType && { fileType }),
                    codeAware: true,
                    preserveCodeSyntax: true,
                    enableUnitConversion: false // Only focus on spellings for diagnostics
                }
            });

            // Create diagnostics from spelling changes
            const diagnostics = this.createDiagnostics(document, response);
            
            // Update diagnostic collection
            this.diagnosticCollection.set(document.uri, diagnostics);
            
            this.logDebug(`Processed ${document.fileName}: found ${diagnostics.length} American spellings`);
            
        } catch {
            const message = "An unknown error occurred";
            this.outputChannel.appendLine(`[Diagnostics] Error processing ${document.fileName}: ${message}`);
            
            // Clear diagnostics on error
            this.diagnosticCollection.set(document.uri, []);
        }
    }

    /**
     * Create diagnostic entries from conversion response
     */
    private createDiagnostics(document: vscode.TextDocument, response: ConvertResponse): vscode.Diagnostic[] {
        const config = vscode.workspace.getConfiguration('m2e');
        const severity = this.getSeverityFromConfig(config.get<string>('diagnosticSeverity', 'Information'));
        
        const diagnostics: vscode.Diagnostic[] = [];
        
        // Only process spelling changes (ignore unit conversions for diagnostics)
        const spellingChanges = response.changes.filter(change => change.type === 'spelling');
        
        for (const change of spellingChanges) {
            // Skip if word is in ignore list
            if (this.workspaceIgnoreList.has(change.original.toLowerCase())) {
                continue;
            }
            
            try {
                const startPos = document.positionAt(change.position);
                const endPos = document.positionAt(change.position + change.original.length);
                const range = new vscode.Range(startPos, endPos);
                
                const diagnostic = new vscode.Diagnostic(
                    range,
                    `American spelling: "${change.original}" → suggest "${change.converted}"`,
                    severity
                );
                
                diagnostic.source = 'M2E';
                diagnostic.code = 'american-spelling';
                
                // Add related information
                diagnostic.relatedInformation = [
                    new vscode.DiagnosticRelatedInformation(
                        new vscode.Location(document.uri, range),
                        `British spelling: ${change.converted}`
                    )
                ];

                diagnostics.push(diagnostic);
                
            } catch {
                // Skip invalid positions
                this.logDebug(`Invalid position for change: ${change.original} at ${change.position}`);
            }
        }
        
        return diagnostics;
    }

    /**
     * Convert configuration string to VSCode diagnostic severity
     */
    private getSeverityFromConfig(severityString: string): vscode.DiagnosticSeverity {
        switch (severityString.toLowerCase()) {
            case 'error':
                return vscode.DiagnosticSeverity.Error;
            case 'warning':
                return vscode.DiagnosticSeverity.Warning;
            case 'information':
            default:
                return vscode.DiagnosticSeverity.Information;
        }
    }

    /**
     * Clear diagnostics for a document
     */
    private clearDocumentDiagnostics(document: vscode.TextDocument): void {
        const uri = document.uri.toString();
        this.diagnosticCollection.delete(document.uri);
        this.lastProcessedVersion.delete(uri);
        
        // Clear any pending timers
        const timer = this.updateTimers.get(uri);
        if (timer) {
            clearTimeout(timer);
            this.updateTimers.delete(uri);
        }
    }

    /**
     * Get diagnostics for a document
     */
    public getDiagnostics(document: vscode.TextDocument): readonly vscode.Diagnostic[] {
        return this.diagnosticCollection.get(document.uri) || [];
    }

    /**
     * Add word to workspace ignore list
     */
    public async addToIgnoreList(word: string): Promise<void> {
        const normalizedWord = word.toLowerCase();
        this.workspaceIgnoreList.add(normalizedWord);
        
        await this.saveWorkspaceIgnoreList();
        
        // Refresh diagnostics for all open documents
        this.processOpenDocuments();
        
        this.outputChannel.appendLine(`[Diagnostics] Added "${word}" to ignore list`);
    }

    /**
     * Remove word from workspace ignore list
     */
    public async removeFromIgnoreList(word: string): Promise<void> {
        const normalizedWord = word.toLowerCase();
        this.workspaceIgnoreList.delete(normalizedWord);
        
        await this.saveWorkspaceIgnoreList();
        
        // Refresh diagnostics for all open documents
        this.processOpenDocuments();
        
        this.outputChannel.appendLine(`[Diagnostics] Removed "${word}" from ignore list`);
    }

    /**
     * Check if word is in ignore list
     */
    public isIgnored(word: string): boolean {
        return this.workspaceIgnoreList.has(word.toLowerCase());
    }

    /**
     * Load workspace ignore list from file
     */
    private async loadWorkspaceIgnoreList(): Promise<void> {
        try {
            const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
            if (!workspaceFolder) {
                return;
            }

            const ignoreFileUri = vscode.Uri.joinPath(workspaceFolder.uri, this.ignoreListFile);
            
            try {
                const content = await vscode.workspace.fs.readFile(ignoreFileUri);
                const ignoreData = JSON.parse(content.toString()) as { words: string[] };
                
                this.workspaceIgnoreList.clear();
                if (Array.isArray(ignoreData.words)) {
                    ignoreData.words.forEach(word => {
                        this.workspaceIgnoreList.add(word.toLowerCase());
                    });
                }
                
                this.logDebug(`Loaded ${this.workspaceIgnoreList.size} words from ignore list`);
                
            } catch {
                // File doesn't exist or invalid format - start with empty list
                this.workspaceIgnoreList.clear();
                this.logDebug('No existing ignore list found, starting fresh');
            }
            
        } catch {
            this.outputChannel.appendLine(`[Diagnostics] Error loading ignore list`);
        }
    }

    /**
     * Save workspace ignore list to file
     */
    private async saveWorkspaceIgnoreList(): Promise<void> {
        try {
            const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
            if (!workspaceFolder) {
                return;
            }

            const ignoreFileUri = vscode.Uri.joinPath(workspaceFolder.uri, this.ignoreListFile);
            const ignoreData = {
                words: Array.from(this.workspaceIgnoreList).sort(),
                lastModified: new Date().toISOString(),
                version: '1.0'
            };
            
            const content = JSON.stringify(ignoreData, null, 2);
            await vscode.workspace.fs.writeFile(ignoreFileUri, Buffer.from(content, 'utf8'));
            
            this.logDebug(`Saved ${this.workspaceIgnoreList.size} words to ignore list`);
            
        } catch {
            this.outputChannel.appendLine(`[Diagnostics] Error saving ignore list`);
        }
    }

    /**
     * Force refresh of all diagnostics
     */
    public refreshAll(): void {
        this.lastProcessedVersion.clear();
        this.processOpenDocuments();
    }

    /**
     * Get current workspace ignore list
     */
    public getIgnoreList(): string[] {
        return Array.from(this.workspaceIgnoreList).sort();
    }

    /**
     * Dispose of the diagnostic provider
     */
    public dispose(): void {
        this.diagnosticCollection.dispose();
        
        // Clear all timers
        this.updateTimers.forEach(timer => clearTimeout(timer));
        this.updateTimers.clear();
    }

    /**
     * Debug logging helper
     */
    private logDebug(message: string): void {
        const debugLogging = vscode.workspace.getConfiguration('m2e').get<boolean>('debugLogging', false);
        if (debugLogging) {
            this.outputChannel.appendLine(`[Diagnostics Debug] ${message}`);
        }
    }
}