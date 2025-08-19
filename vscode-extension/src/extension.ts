import * as vscode from 'vscode';
import { M2EServerManager } from './services/server';
import { M2EApiClient } from './services/client';
import { CommandRegistry } from './commands';
import { 
    M2EDiagnosticProvider, 
    M2ECodeActionProvider,
    registerIgnoreWordCommand,
    registerManageIgnoreListCommand,
    registerRefreshDiagnosticsCommand
} from './providers';

let serverManager: M2EServerManager;
let apiClient: M2EApiClient;
let outputChannel: vscode.OutputChannel;
let commandRegistry: CommandRegistry;
let diagnosticProvider: M2EDiagnosticProvider;
let codeActionProvider: M2ECodeActionProvider;

/**
 * Extension activation entry point
 */
export async function activate(context: vscode.ExtensionContext): Promise<void> {
    try {
        // Create output channel for logging
        outputChannel = vscode.window.createOutputChannel('M2E');
        context.subscriptions.push(outputChannel);

        // Initialise server manager and API client
        serverManager = new M2EServerManager(context, outputChannel);
        apiClient = new M2EApiClient(outputChannel);

        // Initialise command registry with server running helper
        commandRegistry = new CommandRegistry(apiClient, outputChannel, ensureServerRunning);

        // Register all commands
        commandRegistry.registerAllCommands(context, serverManager);

        // Initialise diagnostic and code action providers
        diagnosticProvider = new M2EDiagnosticProvider(context, apiClient, outputChannel, ensureServerRunning);
        codeActionProvider = new M2ECodeActionProvider(apiClient, diagnosticProvider, outputChannel, ensureServerRunning);

        // Register code action provider for all supported languages
        const supportedLanguages = [
            'plaintext', 'markdown', 'javascript', 'typescript', 'python', 'go', 
            'java', 'c', 'cpp', 'csharp', 'php', 'ruby', 'rust'
        ];
        
        context.subscriptions.push(
            vscode.languages.registerCodeActionsProvider(
                supportedLanguages,
                codeActionProvider,
                {
                    providedCodeActionKinds: [vscode.CodeActionKind.QuickFix]
                }
            )
        );

        // Register additional commands for diagnostic management
        registerIgnoreWordCommand(context, diagnosticProvider);
        registerManageIgnoreListCommand(context, diagnosticProvider);
        registerRefreshDiagnosticsCommand(context, diagnosticProvider, codeActionProvider);


        outputChannel.appendLine('M2E extension activated successfully');
        

    } catch {
        const message = "An unknown error occurred";
        outputChannel.appendLine(`Failed to activate M2E extension: ${message}`);
        vscode.window.showErrorMessage(`M2E: Failed to activate extension: ${message}`);
    }
}

/**
 * Extension deactivation cleanup
 */
export async function deactivate(): Promise<void> {
    try {
        outputChannel?.appendLine('Deactivating M2E extension...');
        
        // Dispose providers
        if (diagnosticProvider) {
            diagnosticProvider.dispose();
        }
        
        if (codeActionProvider) {
            codeActionProvider.clearCache();
        }
        
        // Stop server gracefully
        if (serverManager) {
            await serverManager.stop();
        }
        
        outputChannel?.appendLine('M2E extension deactivated');
    } catch {
        console.error("Error during M2E extension deactivation");
    }
}

/**
 * Ensure server is running before API calls
 */
async function ensureServerRunning(): Promise<void> {
    const status = serverManager.getStatus();
    
    if (!status.isRunning) {
        outputChannel.appendLine('Starting M2E server...');
        
        const success = await serverManager.start();
        
        if (!success) {
            throw new Error('Failed to start M2E server');
        }
        
        apiClient.setServerUrl(serverManager.getServerUrl());
    }
}




