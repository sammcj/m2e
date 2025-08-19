import * as vscode from 'vscode';
import { ConvertCommands } from './convert';
import { PreviewCommands } from './preview';
import { ReportCommands } from './reports';
import { M2EApiClient } from '../services/client';

/**
 * Command registry and factory for M2E extension commands
 */
export class CommandRegistry {
    private convertCommands: ConvertCommands;
    private previewCommands: PreviewCommands;
    private reportCommands: ReportCommands;

    constructor(
        apiClient: M2EApiClient,
        outputChannel: vscode.OutputChannel,
        ensureServerRunning: () => Promise<void>
    ) {
        this.convertCommands = new ConvertCommands(apiClient, outputChannel, ensureServerRunning);
        this.previewCommands = new PreviewCommands(apiClient, outputChannel, ensureServerRunning);
        this.reportCommands = new ReportCommands(apiClient, outputChannel, ensureServerRunning);
    }

    /**
     * Register all M2E commands with VSCode
     */
    registerAllCommands(context: vscode.ExtensionContext, serverManager: any): void {
        const commands = [
            {
                name: 'm2e.convertSelection',
                handler: () => this.convertCommands.convertSelection(),
                description: 'Convert selected text to British English'
            },
            {
                name: 'm2e.convertFile',
                handler: (uri?: vscode.Uri) => this.convertCommands.convertFile(uri),
                description: 'Convert entire file to British English'
            },
            {
                name: 'm2e.convertCommentsOnly',
                handler: () => this.convertCommands.convertCommentsOnly(),
                description: 'Convert only comments in code files'
            },
            {
                name: 'm2e.convertAndPreview',
                handler: () => this.previewCommands.convertAndPreview(),
                description: 'Preview conversion changes before applying'
            },
            {
                name: 'm2e.restartServer',
                handler: () => this.convertCommands.restartServer(serverManager),
                description: 'Restart the M2E server'
            },
            {
                name: 'm2e.generateReport',
                handler: () => this.reportCommands.generateReport(),
                description: 'Generate M2E analysis report'
            },
            {
                name: 'm2e.convertProject',
                handler: () => this.convertCommands.convertProject(),
                description: 'Convert entire project to British English'
            }
        ];

        // Register each command with VSCode
        commands.forEach(cmd => {
            const disposable = vscode.commands.registerCommand(cmd.name, cmd.handler);
            context.subscriptions.push(disposable);
        });

        // Log successful registration
        const commandNames = commands.map(cmd => cmd.name).join(', ');
        console.log(`M2E: Registered commands: ${commandNames}`);
    }

    /**
     * Get command handlers for external access
     */
    getConvertCommands(): ConvertCommands {
        return this.convertCommands;
    }

    /**
     * Get preview command handlers for external access
     */
    getPreviewCommands(): PreviewCommands {
        return this.previewCommands;
    }
}

// Export command classes for direct access if needed
export { ConvertCommands } from './convert';
export { PreviewCommands } from './preview';
export { ReportCommands } from './reports';