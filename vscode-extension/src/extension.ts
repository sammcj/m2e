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
let statusBarItem: vscode.StatusBarItem;
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

        // Create status bar item
        statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
        context.subscriptions.push(statusBarItem);

        // Initialise server manager and API client
        serverManager = new M2EServerManager(context, outputChannel, statusBarItem);
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

        // Set up configuration change monitoring
        context.subscriptions.push(
            vscode.workspace.onDidChangeConfiguration(onConfigurationChanged)
        );

        outputChannel.appendLine('M2E extension activated successfully');
        
        // Show initial status (server not started yet)
        updateStatusBar('stopped');

    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
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
    } catch (error) {
        console.error('Error during M2E extension deactivation:', error);
    }
}

/**
 * Ensure server is running before API calls
 */
async function ensureServerRunning(): Promise<void> {
    const status = serverManager.getStatus();
    
    if (!status.isRunning) {
        outputChannel.appendLine('Starting M2E server...');
        updateStatusBar('stopped', 'Starting server...');
        
        const success = await serverManager.start();
        
        if (!success) {
            updateStatusBar('error', 'Failed to start');
            throw new Error('Failed to start M2E server');
        }
        
        apiClient.setServerUrl(serverManager.getServerUrl());
        updateStatusBar('running');
    }
}



/**
 * Update status bar display
 */
function updateStatusBar(status: 'running' | 'stopped' | 'error', tooltip?: string): void {
    const showStatusBar = vscode.workspace.getConfiguration('m2e').get<boolean>('showStatusBar', true);
    
    if (!showStatusBar) {
        statusBarItem.hide();
        return;
    }

    const icons = {
        running: '$(circle-filled)',
        stopped: '$(circle-outline)', 
        error: '$(error)'
    };

    const colors = {
        running: undefined, // Default colour
        stopped: new vscode.ThemeColor('statusBarItem.warningBackground'),
        error: new vscode.ThemeColor('statusBarItem.errorBackground')
    };

    statusBarItem.text = `M2E ${icons[status]}`;
    statusBarItem.tooltip = tooltip || `M2E Server is ${status}${status === 'running' ? ` on port ${serverManager?.getStatus().port || ''}` : ''}`;
    statusBarItem.backgroundColor = colors[status];
    statusBarItem.command = status === 'error' ? 'm2e.restartServer' : undefined;
    statusBarItem.show();
}

/**
 * Handle configuration changes
 */
function onConfigurationChanged(event: vscode.ConfigurationChangeEvent): void {
    if (event.affectsConfiguration('m2e.showStatusBar')) {
        const showStatusBar = vscode.workspace.getConfiguration('m2e').get<boolean>('showStatusBar', true);
        if (showStatusBar) {
            const status = serverManager?.getStatus();
            updateStatusBar(status?.isRunning ? 'running' : 'stopped');
        } else {
            statusBarItem.hide();
        }
    }
}