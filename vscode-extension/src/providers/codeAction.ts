import * as vscode from 'vscode';
import { M2EApiClient, ConvertResponse, getFileTypeFromDocument } from '../services/client';
import { M2EDiagnosticProvider } from './diagnostic';

/**
 * Code action provider for M2E Quick Fix suggestions
 */
export class M2ECodeActionProvider implements vscode.CodeActionProvider {
    private apiClient: M2EApiClient;
    private diagnosticProvider: M2EDiagnosticProvider;
    private outputChannel: vscode.OutputChannel;
    private ensureServerRunning: () => Promise<void>;
    
    // Cache for conversion results to avoid repeated API calls
    private conversionCache = new Map<string, ConvertResponse>();
    private cacheTimeout = 30000; // 30 seconds

    constructor(
        apiClient: M2EApiClient,
        diagnosticProvider: M2EDiagnosticProvider,
        outputChannel: vscode.OutputChannel,
        ensureServerRunning: () => Promise<void>
    ) {
        this.apiClient = apiClient;
        this.diagnosticProvider = diagnosticProvider;
        this.outputChannel = outputChannel;
        this.ensureServerRunning = ensureServerRunning;
    }

    /**
     * Provide code actions for M2E diagnostics
     */
    public async provideCodeActions(
        document: vscode.TextDocument,
        _range: vscode.Range | vscode.Selection,
        context: vscode.CodeActionContext,
        _token: vscode.CancellationToken
    ): Promise<vscode.CodeAction[]> {
        const actions: vscode.CodeAction[] = [];

        try {
            // Filter for M2E diagnostics
            const m2eDiagnostics = context.diagnostics.filter(
                diagnostic => diagnostic.source === 'M2E' && diagnostic.code === 'american-spelling'
            );

            // Always try to provide "Fix All" source action, even if no diagnostics in current selection
            if (context.only && context.only.contains(vscode.CodeActionKind.Source)) {
                const sourceActions = await this.createSourceActions(document);
                actions.push(...sourceActions);
            }

            if (m2eDiagnostics.length === 0) {
                return actions;
            }

            // Get conversion data for the document
            const conversionData = await this.getConversionData(document);
            if (!conversionData) {
                return actions;
            }

            // Create actions for each diagnostic
            for (const diagnostic of m2eDiagnostics) {
                const wordActions = this.createWordActions(document, diagnostic, conversionData);
                actions.push(...wordActions);
            }

            // Add document-level actions if any diagnostics
            if (m2eDiagnostics.length > 0) {
                const documentActions = this.createDocumentActions(document, conversionData);
                actions.push(...documentActions);
            }

            // Add ignore actions
            for (const diagnostic of m2eDiagnostics) {
                const ignoreActions = this.createIgnoreActions(document, diagnostic);
                actions.push(...ignoreActions);
            }

        } catch {
            const message = "An unknown error occurred";
            this.outputChannel.appendLine(`[Code Actions] Error providing actions: ${message}`);
        }

        return actions;
    }

    /**
     * Get conversion data for document (with caching)
     */
    private async getConversionData(document: vscode.TextDocument): Promise<ConvertResponse | null> {
        const cacheKey = `${document.uri.toString()}-${document.version}`;
        
        // Check cache first
        const cached = this.conversionCache.get(cacheKey);
        if (cached) {
            return cached;
        }

        try {
            await this.ensureServerRunning();
            
            const text = document.getText();
            const fileType = getFileTypeFromDocument(document);
            
            const response = await this.apiClient.convert({
                text,
                options: {
                    ...(fileType && { fileType }),
                    codeAware: true,
                    preserveCodeSyntax: true,
                    enableUnitConversion: false // Only spellings for code actions
                }
            });

            // Cache the result
            this.conversionCache.set(cacheKey, response);
            
            // Clear cache after timeout
            setTimeout(() => {
                this.conversionCache.delete(cacheKey);
            }, this.cacheTimeout);

            return response;
            
        } catch {
            const message = "An unknown error occurred";
            this.outputChannel.appendLine(`[Code Actions] Failed to get conversion data: ${message}`);
            return null;
        }
    }

    /**
     * Create word-level code actions
     */
    private createWordActions(
        document: vscode.TextDocument,
        diagnostic: vscode.Diagnostic,
        conversionData: ConvertResponse
    ): vscode.CodeAction[] {
        const actions: vscode.CodeAction[] = [];
        
        try {
            // Find the corresponding change in conversion data
            const wordRange = diagnostic.range;
            const originalWord = document.getText(wordRange);
            
            const change = conversionData.changes.find(c => 
                c.type === 'spelling' && 
                c.original === originalWord &&
                this.isPositionInRange(document.positionAt(c.position), wordRange)
            );

            if (!change) {
                return actions;
            }

            // Create "Convert word" action
            const convertWordAction = new vscode.CodeAction(
                `Convert "${change.original}" to "${change.converted}"`,
                vscode.CodeActionKind.QuickFix
            );
            
            convertWordAction.edit = new vscode.WorkspaceEdit();
            convertWordAction.edit.replace(document.uri, wordRange, change.converted);
            
            convertWordAction.diagnostics = [diagnostic];
            convertWordAction.isPreferred = true; // Make this the default action
            
            actions.push(convertWordAction);

        } catch {
            this.logDebug(`Error creating word action`);
        }

        return actions;
    }

    /**
     * Create document-level code actions
     */
    private createDocumentActions(
        document: vscode.TextDocument,
        conversionData: ConvertResponse
    ): vscode.CodeAction[] {
        const actions: vscode.CodeAction[] = [];
        
        try {
            const spellingChanges = conversionData.changes.filter(c => c.type === 'spelling');
            
            if (spellingChanges.length === 0) {
                return actions;
            }

            // Create "Convert all American spellings" action
            const convertAllAction = new vscode.CodeAction(
                `Convert all American spellings (${spellingChanges.length} changes)`,
                vscode.CodeActionKind.QuickFix
            );
            
            convertAllAction.edit = new vscode.WorkspaceEdit();
            
            // Apply all spelling changes (in reverse order to maintain positions)
            const sortedChanges = [...spellingChanges].sort((a, b) => b.position - a.position);
            
            for (const change of sortedChanges) {
                try {
                    const startPos = document.positionAt(change.position);
                    const endPos = document.positionAt(change.position + change.original.length);
                    const range = new vscode.Range(startPos, endPos);
                    
                    // Skip if word is in ignore list
                    if (this.diagnosticProvider.isIgnored(change.original)) {
                        continue;
                    }
                    
                    convertAllAction.edit.replace(document.uri, range, change.converted);
                } catch {
                    this.logDebug(`Skipping invalid position: ${change.position}`);
                }
            }
            
            actions.push(convertAllAction);

        } catch {
            this.logDebug(`Error creating document actions`);
        }

        return actions;
    }

    /**
     * Create source actions (appear in "Source Action" menu)
     */
    private async createSourceActions(document: vscode.TextDocument): Promise<vscode.CodeAction[]> {
        const actions: vscode.CodeAction[] = [];
        
        try {
            // Get conversion data to see if there are American spellings to fix
            const conversionData = await this.getConversionData(document);
            if (!conversionData) {
                return actions;
            }

            const spellingChanges = conversionData.changes.filter(c => c.type === 'spelling');
            
            if (spellingChanges.length > 0) {
                // Create "Fix All American Spellings" source action
                const fixAllAction = new vscode.CodeAction(
                    `Fix All American Spellings (${spellingChanges.length} changes)`,
                    vscode.CodeActionKind.SourceFixAll
                );
                
                // Use command instead of direct edit for better UX
                fixAllAction.command = {
                    title: 'Fix All American Spellings',
                    command: 'm2e.fixAllAmericanisations'
                };
                
                fixAllAction.isPreferred = true;
                actions.push(fixAllAction);
            }

        } catch {
            this.logDebug(`Error creating source actions`);
        }

        return actions;
    }

    /**
     * Create ignore-related code actions
     */
    private createIgnoreActions(
        document: vscode.TextDocument,
        diagnostic: vscode.Diagnostic
    ): vscode.CodeAction[] {
        const actions: vscode.CodeAction[] = [];
        
        try {
            const wordRange = diagnostic.range;
            const originalWord = document.getText(wordRange);
            
            // Create "Ignore word" action
            const ignoreWordAction = new vscode.CodeAction(
                `Ignore "${originalWord}" in this workspace`,
                vscode.CodeActionKind.QuickFix
            );
            
            ignoreWordAction.command = {
                title: 'Ignore word',
                command: 'm2e.ignoreWord',
                arguments: [originalWord]
            };
            
            ignoreWordAction.diagnostics = [diagnostic];
            
            actions.push(ignoreWordAction);

        } catch {
            this.logDebug(`Error creating ignore actions`);
        }

        return actions;
    }

    /**
     * Check if position is within range
     */
    private isPositionInRange(position: vscode.Position, range: vscode.Range): boolean {
        return range.contains(position);
    }

    /**
     * Clear conversion cache
     */
    public clearCache(): void {
        this.conversionCache.clear();
    }

    /**
     * Debug logging helper
     */
    private logDebug(message: string): void {
        const debugLogging = vscode.workspace.getConfiguration('m2e').get<boolean>('debugLogging', false);
        if (debugLogging) {
            this.outputChannel.appendLine(`[Code Actions Debug] ${message}`);
        }
    }
}

/**
 * Helper function to register ignore word command
 */
export function registerIgnoreWordCommand(
    context: vscode.ExtensionContext,
    diagnosticProvider: M2EDiagnosticProvider
): void {
    const disposable = vscode.commands.registerCommand('m2e.ignoreWord', async (word: string) => {
        try {
            await diagnosticProvider.addToIgnoreList(word);
            vscode.window.showInformationMessage(`M2E: Added "${word}" to workspace ignore list`);
        } catch {
            const message = "An unknown error occurred";
            vscode.window.showErrorMessage(`M2E: Failed to ignore word: ${message}`);
        }
    });
    
    context.subscriptions.push(disposable);
}

/**
 * Helper function to register manage ignore list command
 */
export function registerManageIgnoreListCommand(
    context: vscode.ExtensionContext,
    diagnosticProvider: M2EDiagnosticProvider
): void {
    const disposable = vscode.commands.registerCommand('m2e.manageIgnoreList', async () => {
        try {
            const ignoreList = diagnosticProvider.getIgnoreList();
            
            if (ignoreList.length === 0) {
                vscode.window.showInformationMessage('M2E: No words in ignore list');
                return;
            }

            const selected = await vscode.window.showQuickPick(
                ignoreList.map(word => ({
                    label: word,
                    description: 'Remove from ignore list'
                })),
                {
                    title: 'M2E Ignore List',
                    placeHolder: 'Select word to remove from ignore list'
                }
            );

            if (selected) {
                await diagnosticProvider.removeFromIgnoreList(selected.label);
                vscode.window.showInformationMessage(`M2E: Removed "${selected.label}" from ignore list`);
            }

        } catch {
            const message = "An unknown error occurred";
            vscode.window.showErrorMessage(`M2E: Failed to manage ignore list: ${message}`);
        }
    });
    
    context.subscriptions.push(disposable);
}

/**
 * Helper function to register refresh diagnostics command
 */
export function registerRefreshDiagnosticsCommand(
    context: vscode.ExtensionContext,
    diagnosticProvider: M2EDiagnosticProvider,
    codeActionProvider: M2ECodeActionProvider
): void {
    const disposable = vscode.commands.registerCommand('m2e.refreshDiagnostics', () => {
        try {
            codeActionProvider.clearCache();
            diagnosticProvider.refreshAll();
            vscode.window.showInformationMessage('M2E: Refreshed diagnostics for all open documents');
        } catch {
            const message = "An unknown error occurred";
            vscode.window.showErrorMessage(`M2E: Failed to refresh diagnostics: ${message}`);
        }
    });
    
    context.subscriptions.push(disposable);
}