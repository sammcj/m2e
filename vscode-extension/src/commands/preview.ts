import * as vscode from 'vscode';
import { M2EApiClient, getFileTypeFromDocument, ConvertResponse } from '../services/client';

/**
 * Preview and diff functionality for the M2E extension
 */
export class PreviewCommands {
    constructor(
        private apiClient: M2EApiClient,
        private outputChannel: vscode.OutputChannel,
        private ensureServerRunning: () => Promise<void>
    ) {}

    /**
     * Convert and show preview with diff comparison
     */
    async convertAndPreview(): Promise<void> {
        try {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showWarningMessage('M2E: No active editor found');
                return;
            }

            const document = editor.document;
            let textToConvert: string;
            let isSelection = false;
            let range: vscode.Range | undefined;

            // Determine what to convert: selection or entire document
            if (!editor.selection.isEmpty) {
                textToConvert = document.getText(editor.selection);
                isSelection = true;
                range = editor.selection;
            } else {
                textToConvert = document.getText();
                range = new vscode.Range(
                    document.positionAt(0),
                    document.positionAt(textToConvert.length)
                );
            }

            if (!textToConvert.trim()) {
                vscode.window.showWarningMessage('M2E: No text to convert');
                return;
            }

            // Validate text size for performance
            if (textToConvert.length > 200000) { // 200KB limit for preview
                const proceed = await vscode.window.showWarningMessage(
                    'M2E: Large text detected. Preview may take some time to generate.',
                    'Continue',
                    'Cancel'
                );
                if (proceed !== 'Continue') {
                    return;
                }
            }

            await this.ensureServerRunning();

            const fileType = getFileTypeFromDocument(document);
            this.outputChannel.appendLine(
                `Generating preview for ${isSelection ? 'selection' : 'file'} (type: ${fileType}, ${textToConvert.length} characters)`
            );

            // Show progress for preview generation
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: `M2E: Generating preview...`,
                cancellable: false
            }, async () => {
                try {
                    const result = await this.apiClient.convertFile(textToConvert, fileType);
                    await this.showConversionPreview(result, document, isSelection, range);
                } catch (error) {
                    throw error; // Re-throw to be caught by outer try-catch
                }
            });

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Convert and preview failed: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to generate preview: ${message}`);
        }
    }

    /**
     * Show conversion preview in diff editor with detailed change information
     */
    private async showConversionPreview(
        result: ConvertResponse,
        document: vscode.TextDocument,
        isSelection: boolean,
        range?: vscode.Range
    ): Promise<void> {
        try {
            const fileName = document.fileName.split('/').pop() || 'Untitled';
            const scope = isSelection ? ' (Selection)' : '';
            
            // Create diff preview
            await this.createDiffPreview(
                result.originalText,
                result.convertedText,
                `M2E Preview${scope} - ${fileName}`,
                document.languageId
            );

            // Show statistics and change summary
            const changeCount = result.metadata.spellingChanges + result.metadata.unitChanges;
            const stats = this.formatChangeStats(result.metadata);
            
            if (changeCount > 0) {
                // Show detailed change information
                const message = `M2E Preview: ${changeCount} potential change${changeCount !== 1 ? 's' : ''} found ${stats}`;
                
                const action = await vscode.window.showInformationMessage(
                    message,
                    'Apply Changes',
                    'Show Change Details',
                    'Dismiss'
                );

                if (action === 'Apply Changes') {
                    await this.applyPreviewedChanges(document, result, isSelection, range);
                } else if (action === 'Show Change Details') {
                    await this.showChangeDetails(result.changes, document.fileName);
                }
            } else {
                vscode.window.showInformationMessage('M2E Preview: No changes would be made');
            }

            this.outputChannel.appendLine(
                `Preview generated: ${stats}, processing time: ${result.metadata.processingTimeMs}ms`
            );

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Failed to show preview: ${message}`);
            throw new Error(`Failed to show preview: ${message}`);
        }
    }

    /**
     * Create and display diff preview in VSCode
     */
    private async createDiffPreview(
        original: string,
        converted: string,
        title: string,
        languageId: string = 'plaintext'
    ): Promise<void> {
        // Create URIs for temporary documents
        const originalUri = vscode.Uri.parse(`untitled:Original - ${title}`).with({
            scheme: 'untitled'
        });
        const convertedUri = vscode.Uri.parse(`untitled:Converted - ${title}`).with({
            scheme: 'untitled'
        });

        try {
            // Create temporary documents with content
            const originalDoc = await vscode.workspace.openTextDocument({
                content: original,
                language: languageId
            });
            const convertedDoc = await vscode.workspace.openTextDocument({
                content: converted,
                language: languageId
            });

            // Show diff view
            await vscode.commands.executeCommand(
                'vscode.diff',
                originalDoc.uri,
                convertedDoc.uri,
                title,
                {
                    preview: true,
                    preserveFocus: false
                }
            );

        } catch (error) {
            // Fallback method for older VSCode versions
            const originalDoc = await vscode.workspace.openTextDocument(originalUri);
            const convertedDoc = await vscode.workspace.openTextDocument(convertedUri);

            // Insert content into documents
            const originalEditor = await vscode.window.showTextDocument(originalDoc, { preview: true });
            await originalEditor.edit(editBuilder => {
                editBuilder.insert(new vscode.Position(0, 0), original);
            });

            const convertedEditor = await vscode.window.showTextDocument(convertedDoc, { preview: true });
            await convertedEditor.edit(editBuilder => {
                editBuilder.insert(new vscode.Position(0, 0), converted);
            });

            // Show diff
            await vscode.commands.executeCommand(
                'vscode.diff',
                originalDoc.uri,
                convertedDoc.uri,
                title
            );
        }
    }

    /**
     * Apply previewed changes to the original document
     */
    private async applyPreviewedChanges(
        document: vscode.TextDocument,
        result: ConvertResponse,
        isSelection: boolean,
        range?: vscode.Range
    ): Promise<void> {
        try {
            // Get or create editor for the document
            const editor = await vscode.window.showTextDocument(document);

            const success = await editor.edit(editBuilder => {
                if (isSelection && range) {
                    editBuilder.replace(range, result.convertedText);
                } else {
                    const entireRange = new vscode.Range(
                        document.positionAt(0),
                        document.positionAt(result.originalText.length)
                    );
                    editBuilder.replace(entireRange, result.convertedText);
                }
            });

            if (success) {
                const changeCount = result.metadata.spellingChanges + result.metadata.unitChanges;
                const scope = isSelection ? 'selection' : document.fileName;
                const stats = this.formatChangeStats(result.metadata);
                
                vscode.window.showInformationMessage(
                    `M2E: Applied ${changeCount} change${changeCount !== 1 ? 's' : ''} to ${scope} ${stats}`
                );

                this.outputChannel.appendLine(`Changes applied successfully to ${scope}`);
            } else {
                vscode.window.showErrorMessage('M2E: Failed to apply changes to document');
            }

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Failed to apply previewed changes: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to apply changes: ${message}`);
        }
    }

    /**
     * Show detailed information about changes found
     */
    private async showChangeDetails(changes: ConvertResponse['changes'], fileName: string): Promise<void> {
        try {
            const changesByType = this.groupChangesByType(changes);
            const content = this.generateChangeDetailsContent(changesByType, fileName);

            // Create and show change details document
            const doc = await vscode.workspace.openTextDocument({
                content,
                language: 'markdown'
            });

            await vscode.window.showTextDocument(doc, {
                preview: true,
                viewColumn: vscode.ViewColumn.Beside
            });

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Failed to show change details: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to show change details: ${message}`);
        }
    }

    /**
     * Group changes by type for better organisation
     */
    private groupChangesByType(changes: ConvertResponse['changes']): {
        spelling: ConvertResponse['changes'];
        unit: ConvertResponse['changes'];
    } {
        const spelling = changes.filter(change => change.type === 'spelling');
        const unit = changes.filter(change => change.type === 'unit');
        
        return { spelling, unit };
    }

    /**
     * Generate markdown content for change details
     */
    private generateChangeDetailsContent(
        changesByType: { spelling: ConvertResponse['changes']; unit: ConvertResponse['changes'] },
        fileName: string
    ): string {
        const lines: string[] = [
            `# M2E Conversion Details - ${fileName}`,
            '',
            `## Summary`,
            `- **Spelling Changes**: ${changesByType.spelling.length}`,
            `- **Unit Conversions**: ${changesByType.unit.length}`,
            `- **Total Changes**: ${changesByType.spelling.length + changesByType.unit.length}`,
            ''
        ];

        if (changesByType.spelling.length > 0) {
            lines.push('## Spelling Changes', '');
            changesByType.spelling.forEach((change, index) => {
                lines.push(`${index + 1}. \`${change.original}\` → \`${change.converted}\``);
            });
            lines.push('');
        }

        if (changesByType.unit.length > 0) {
            lines.push('## Unit Conversions', '');
            changesByType.unit.forEach((change, index) => {
                lines.push(`${index + 1}. \`${change.original}\` → \`${change.converted}\``);
            });
            lines.push('');
        }

        if (changesByType.spelling.length === 0 && changesByType.unit.length === 0) {
            lines.push('## No Changes', '', 'No conversions were found in the analysed text.');
        }

        return lines.join('\n');
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
     * Utility method to create diff ranges for highlighting changes
     */
    static createDiffRanges(changes: ConvertResponse['changes'], document: vscode.TextDocument): vscode.Range[] {
        const ranges: vscode.Range[] = [];
        
        for (const change of changes) {
            try {
                const position = document.positionAt(change.position);
                const endPosition = document.positionAt(change.position + change.original.length);
                ranges.push(new vscode.Range(position, endPosition));
            } catch (error) {
                // Skip invalid positions - they might be beyond document bounds
                console.warn('Invalid position in change:', change, error);
            }
        }
        
        return ranges;
    }
}