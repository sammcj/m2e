import * as vscode from 'vscode';
import { M2EApiClient, getFileTypeFromDocument, ConvertResponse } from '../services/client';

/**
 * Core conversion commands for the M2E extension
 */
export class ConvertCommands {
    constructor(
        private apiClient: M2EApiClient,
        private outputChannel: vscode.OutputChannel,
        private ensureServerRunning: () => Promise<void>
    ) {}

    /**
     * Convert selected text in the active editor
     */
    async convertSelection(): Promise<void> {
        try {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showWarningMessage('M2E: No active editor found');
                return;
            }

            const selection = editor.selection;
            if (selection.isEmpty) {
                vscode.window.showWarningMessage('M2E: No text selected. Please select text to convert.');
                return;
            }

            const selectedText = editor.document.getText(selection);
            if (!selectedText.trim()) {
                vscode.window.showWarningMessage('M2E: Selected text is empty or contains only whitespace');
                return;
            }

            // Validate selection size for performance
            if (selectedText.length > 100000) { // 100KB limit
                const proceed = await vscode.window.showWarningMessage(
                    'M2E: Large selection detected. This may take some time to process.',
                    'Continue',
                    'Cancel'
                );
                if (proceed !== 'Continue') {
                    return;
                }
            }

            await this.ensureServerRunning();

            const fileType = getFileTypeFromDocument(editor.document);
            this.outputChannel.appendLine(`Converting selection in ${editor.document.fileName} (type: ${fileType})`);

            // Show progress for longer operations
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: 'M2E: Converting selection...',
                cancellable: false
            }, async () => {
                try {
                    const result = await this.apiClient.convertSelection(selectedText, fileType);
                    await this.applySelectionConversion(editor, selection, result);
                } catch (error) {
                    throw error; // Re-throw to be caught by outer try-catch
                }
            });

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Convert selection failed: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to convert selection: ${message}`);
        }
    }

    /**
     * Convert entire file content
     */
    async convertFile(uri?: vscode.Uri): Promise<void> {
        try {
            let document: vscode.TextDocument;

            if (uri) {
                // Called from explorer context menu
                try {
                    document = await vscode.workspace.openTextDocument(uri);
                } catch (error) {
                    vscode.window.showErrorMessage(`M2E: Cannot open file: ${error instanceof Error ? error.message : String(error)}`);
                    return;
                }
            } else {
                // Called from command palette or keyboard shortcut
                const editor = vscode.window.activeTextEditor;
                if (!editor) {
                    vscode.window.showWarningMessage('M2E: No active editor found. Please open a file first.');
                    return;
                }
                document = editor.document;
            }

            // Check if document has unsaved changes
            if (document.isDirty) {
                const save = await vscode.window.showWarningMessage(
                    'M2E: File has unsaved changes. Save before converting?',
                    'Save and Continue',
                    'Continue Without Saving',
                    'Cancel'
                );
                
                if (save === 'Cancel') {
                    return;
                } else if (save === 'Save and Continue') {
                    await document.save();
                }
            }

            const fileContent = document.getText();
            if (!fileContent.trim()) {
                vscode.window.showWarningMessage('M2E: File is empty');
                return;
            }

            // Validate file size for performance
            if (fileContent.length > 500000) { // 500KB limit
                const proceed = await vscode.window.showWarningMessage(
                    'M2E: Large file detected. Processing may take some time.',
                    'Continue',
                    'Cancel'
                );
                if (proceed !== 'Continue') {
                    return;
                }
            }

            await this.ensureServerRunning();

            const fileType = getFileTypeFromDocument(document);
            this.outputChannel.appendLine(`Converting file ${document.fileName} (type: ${fileType}, ${fileContent.length} characters)`);

            // Show progress for file conversion
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: `M2E: Converting ${document.fileName}...`,
                cancellable: false
            }, async () => {
                try {
                    const result = await this.apiClient.convertFile(fileContent, fileType);
                    await this.applyFileConversion(document, result, fileContent.length);
                } catch (error) {
                    throw error; // Re-throw to be caught by outer try-catch
                }
            });

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Convert file failed: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to convert file: ${message}`);
        }
    }

    /**
     * Convert only comments in code files
     */
    async convertCommentsOnly(): Promise<void> {
        try {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showWarningMessage('M2E: No active editor found');
                return;
            }

            const document = editor.document;
            const fileContent = document.getText();
            if (!fileContent.trim()) {
                vscode.window.showWarningMessage('M2E: File is empty');
                return;
            }

            const fileType = getFileTypeFromDocument(document);
            if (!fileType || fileType === 'text' || fileType === 'markdown') {
                vscode.window.showWarningMessage('M2E: Comments-only conversion requires a recognised programming language file type');
                return;
            }

            // Check if document has unsaved changes
            if (document.isDirty) {
                const save = await vscode.window.showWarningMessage(
                    'M2E: File has unsaved changes. Save before converting comments?',
                    'Save and Continue',
                    'Continue Without Saving',
                    'Cancel'
                );
                
                if (save === 'Cancel') {
                    return;
                } else if (save === 'Save and Continue') {
                    await document.save();
                }
            }

            await this.ensureServerRunning();

            this.outputChannel.appendLine(`Converting comments only in ${document.fileName} (type: ${fileType})`);

            // Show progress for comment conversion
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: 'M2E: Converting comments...',
                cancellable: false
            }, async () => {
                try {
                    const result = await this.apiClient.convertCommentsOnly(fileContent, fileType);
                    await this.applyFileConversion(document, result, fileContent.length, true);
                } catch (error) {
                    throw error; // Re-throw to be caught by outer try-catch
                }
            });

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Convert comments only failed: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to convert comments: ${message}`);
        }
    }

    /**
     * Restart the M2E server
     */
    async restartServer(serverManager: any): Promise<void> {
        try {
            this.outputChannel.appendLine('Restarting M2E server...');
            
            const success = await serverManager.restart();
            
            if (success) {
                this.apiClient.setServerUrl(serverManager.getServerUrl());
                vscode.window.showInformationMessage('M2E: Server restarted successfully');
                this.outputChannel.appendLine('Server restart completed successfully');
            } else {
                vscode.window.showErrorMessage('M2E: Failed to restart server');
                this.outputChannel.appendLine('Server restart failed');
            }

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Restart server failed: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to restart server: ${message}`);
        }
    }

    /**
     * Apply conversion results to selected text
     */
    private async applySelectionConversion(
        editor: vscode.TextEditor,
        selection: vscode.Selection,
        result: ConvertResponse
    ): Promise<void> {
        if (result.convertedText !== result.originalText) {
            const success = await editor.edit(editBuilder => {
                editBuilder.replace(selection, result.convertedText);
            });

            if (success) {
                const changeCount = result.metadata.spellingChanges + result.metadata.unitChanges;
                const stats = this.formatChangeStats(result.metadata);
                
                vscode.window.showInformationMessage(
                    `M2E: Converted ${changeCount} item${changeCount !== 1 ? 's' : ''} in selection ${stats}`
                );
                
                this.outputChannel.appendLine(
                    `Selection conversion completed: ${stats}, processing time: ${result.metadata.processingTimeMs}ms`
                );
            } else {
                vscode.window.showErrorMessage('M2E: Failed to apply changes to editor');
            }
        } else {
            vscode.window.showInformationMessage('M2E: No changes needed in selection');
            this.outputChannel.appendLine('Selection conversion completed: no changes needed');
        }
    }

    /**
     * Apply conversion results to entire file
     */
    private async applyFileConversion(
        document: vscode.TextDocument,
        result: ConvertResponse,
        originalLength: number,
        commentsOnly: boolean = false
    ): Promise<void> {
        if (result.convertedText !== result.originalText) {
            // Open the document in editor if not already open
            const editor = await vscode.window.showTextDocument(document);
            
            // Replace entire document content
            const success = await editor.edit(editBuilder => {
                const entireRange = new vscode.Range(
                    document.positionAt(0),
                    document.positionAt(originalLength)
                );
                editBuilder.replace(entireRange, result.convertedText);
            });

            if (success) {
                const changeCount = result.metadata.spellingChanges + result.metadata.unitChanges;
                const stats = this.formatChangeStats(result.metadata);
                const scope = commentsOnly ? 'comments' : document.fileName;
                
                vscode.window.showInformationMessage(
                    `M2E: Converted ${changeCount} item${changeCount !== 1 ? 's' : ''} in ${scope} ${stats}`
                );
                
                this.outputChannel.appendLine(
                    `File conversion completed: ${stats}, processing time: ${result.metadata.processingTimeMs}ms`
                );
            } else {
                vscode.window.showErrorMessage('M2E: Failed to apply changes to editor');
            }
        } else {
            const scope = commentsOnly ? 'comments' : document.fileName;
            vscode.window.showInformationMessage(`M2E: No changes needed in ${scope}`);
            this.outputChannel.appendLine(`File conversion completed: no changes needed`);
        }
    }

    /**
     * Format change statistics for display
     */
    private formatChangeStats(metadata: ConvertResponse['metadata']): string {
        const parts: string[] = [];
        
        if (metadata.spellingChanges > 0) {
            parts.push(`${metadata.spellingChanges} spelling`);
        }
        
        if (metadata.unitChanges > 0) {
            parts.push(`${metadata.unitChanges} units`);
        }
        
        return parts.length > 0 ? `(${parts.join(', ')})` : '';
    }

    /**
     * Convert all text files in the current project
     */
    async convertProject(): Promise<void> {
        try {
            const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
            if (!workspaceFolder) {
                vscode.window.showWarningMessage('M2E: No workspace folder found');
                return;
            }

            // Warning about project-wide changes
            const proceed = await vscode.window.showWarningMessage(
                'M2E: This will convert American spellings to British English in ALL text files in your project. This action cannot be undone. Are you sure?',
                { modal: true },
                'Yes, Convert Project',
                'Cancel'
            );

            if (proceed !== 'Yes, Convert Project') {
                return;
            }

            await this.ensureServerRunning();

            // Find text files in the workspace
            const textFiles = await this.findProjectTextFiles(workspaceFolder.uri);
            
            if (textFiles.length === 0) {
                vscode.window.showInformationMessage('M2E: No text files found in workspace');
                return;
            }

            this.outputChannel.appendLine(`Converting ${textFiles.length} files in project`);

            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: `M2E: Converting project files...`,
                cancellable: false
            }, async (progress) => {
                let convertedFiles = 0;
                let totalChanges = 0;
                const errors: string[] = [];

                for (let i = 0; i < textFiles.length; i++) {
                    const file = textFiles[i];
                    const relativePath = vscode.workspace.asRelativePath(file);
                    
                    progress.report({
                        increment: (100 / textFiles.length),
                        message: `Converting ${relativePath}...`
                    });

                    try {
                        const document = await vscode.workspace.openTextDocument(file);
                        const originalText = document.getText();
                        
                        if (originalText.trim()) {
                            const fileType = getFileTypeFromDocument(document);
                            const result = await this.apiClient.convertFile(originalText, fileType);
                            const changeCount = result.metadata.spellingChanges + result.metadata.unitChanges;
                            
                            // Only write file if there are changes
                            if (changeCount > 0 && result.convertedText !== originalText) {
                                const edit = new vscode.WorkspaceEdit();
                                const fullRange = new vscode.Range(
                                    document.positionAt(0),
                                    document.positionAt(originalText.length)
                                );
                                edit.replace(file, fullRange, result.convertedText);
                                await vscode.workspace.applyEdit(edit);
                                
                                convertedFiles++;
                                totalChanges += changeCount;
                                
                                this.outputChannel.appendLine(
                                    `Converted ${relativePath}: ${changeCount} changes`
                                );
                            }
                        }
                    } catch (error) {
                        const errorMessage = error instanceof Error ? error.message : String(error);
                        errors.push(`${relativePath}: ${errorMessage}`);
                        this.outputChannel.appendLine(`Error converting ${relativePath}: ${errorMessage}`);
                    }
                }

                // Show completion summary
                let message = `M2E: Project conversion complete! `;
                message += `${convertedFiles} file${convertedFiles !== 1 ? 's' : ''} converted, `;
                message += `${totalChanges} total change${totalChanges !== 1 ? 's' : ''} made`;

                if (errors.length > 0) {
                    message += ` (${errors.length} error${errors.length !== 1 ? 's' : ''})`;
                }

                vscode.window.showInformationMessage(message);
                
                this.outputChannel.appendLine(
                    `Project conversion completed: ${convertedFiles} files converted, ${totalChanges} changes made`
                );

                if (errors.length > 0) {
                    this.outputChannel.appendLine(`Errors encountered in ${errors.length} files:`);
                    errors.forEach(error => this.outputChannel.appendLine(`  ${error}`));
                }
            });

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Convert project failed: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to convert project: ${message}`);
        }
    }

    /**
     * Find text files in the project workspace
     */
    private async findProjectTextFiles(_workspaceUri: vscode.Uri): Promise<vscode.Uri[]> {
        const config = vscode.workspace.getConfiguration('m2e');
        const excludePatterns = config.get<string[]>('excludePatterns', []);
        
        // Common text file patterns
        const includePatterns = [
            '**/*.txt',
            '**/*.md',
            '**/*.js',
            '**/*.ts',
            '**/*.jsx',
            '**/*.tsx',
            '**/*.py',
            '**/*.go',
            '**/*.java',
            '**/*.c',
            '**/*.cpp',
            '**/*.h',
            '**/*.hpp',
            '**/*.cs',
            '**/*.php',
            '**/*.rb',
            '**/*.rs',
            '**/*.json',
            '**/*.xml',
            '**/*.html',
            '**/*.css',
            '**/*.scss',
            '**/*.less'
        ];

        const files: vscode.Uri[] = [];
        
        for (const pattern of includePatterns) {
            try {
                const foundFiles = await vscode.workspace.findFiles(
                    pattern,
                    `{${excludePatterns.join(',')}}`,
                    1000 // Limit to 1000 files for performance
                );
                files.push(...foundFiles);
            } catch (error) {
                // Continue with other patterns if one fails
            }
        }

        // Remove duplicates
        const uniqueFiles = Array.from(new Set(files.map(f => f.toString())))
            .map(uriString => vscode.Uri.parse(uriString));

        return uniqueFiles.slice(0, 100); // Limit to 100 files for performance
    }
}