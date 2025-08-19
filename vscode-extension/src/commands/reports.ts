import * as vscode from 'vscode';
import * as path from 'path';
import { M2EApiClient, getFileTypeFromDocument } from '../services/client';

/**
 * Report generation commands for the M2E extension
 */
export class ReportCommands {
    constructor(
        private apiClient: M2EApiClient,
        private outputChannel: vscode.OutputChannel,
        private ensureServerRunning: () => Promise<void>
    ) {}

    /**
     * Generate a report for the current file or project
     */
    async generateReport(): Promise<void> {
        try {
            const options = ['Current File', 'Entire Project'];
            const choice = await vscode.window.showQuickPick(options, {
                placeHolder: 'Generate report for...'
            });

            if (!choice) {
                return;
            }

            if (choice === 'Current File') {
                await this.generateFileReport();
            } else {
                await this.generateProjectReport();
            }

        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            this.outputChannel.appendLine(`Generate report failed: ${message}`);
            vscode.window.showErrorMessage(`M2E: Failed to generate report: ${message}`);
        }
    }

    /**
     * Generate report for the current file
     */
    private async generateFileReport(): Promise<void> {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            vscode.window.showWarningMessage('M2E: No active editor found');
            return;
        }

        const document = editor.document;
        const text = document.getText();

        if (!text.trim()) {
            vscode.window.showWarningMessage('M2E: No text to analyse');
            return;
        }

        await this.ensureServerRunning();

        const fileType = getFileTypeFromDocument(document);
        const fileName = path.basename(document.fileName);

        this.outputChannel.appendLine(`Generating report for ${fileName} (type: ${fileType})`);

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: `M2E: Generating report for ${fileName}...`,
            cancellable: false
        }, async () => {
            const result = await this.apiClient.convertFile(text, fileType);
            await this.showReportInNewTab(result, fileName, 'File Report');
        });
    }

    /**
     * Generate report for the entire project
     */
    private async generateProjectReport(): Promise<void> {
        const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
        if (!workspaceFolder) {
            vscode.window.showWarningMessage('M2E: No workspace folder found');
            return;
        }

        await this.ensureServerRunning();

        // Find text files in the workspace
        const textFiles = await this.findTextFiles(workspaceFolder.uri);
        
        if (textFiles.length === 0) {
            vscode.window.showInformationMessage('M2E: No text files found in workspace');
            return;
        }

        this.outputChannel.appendLine(`Generating project report for ${textFiles.length} files`);

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: `M2E: Generating project report...`,
            cancellable: false
        }, async (progress) => {
            const projectResults: Array<{
                fileName: string;
                relativePath: string;
                result: any;
                wordCount: number;
                changeCount: number;
            }> = [];

            let totalChanges = 0;
            let totalWords = 0;

            for (let i = 0; i < textFiles.length; i++) {
                const file = textFiles[i];
                const relativePath = vscode.workspace.asRelativePath(file);
                
                progress.report({
                    increment: (100 / textFiles.length),
                    message: `Processing ${relativePath}...`
                });

                try {
                    const document = await vscode.workspace.openTextDocument(file);
                    const text = document.getText();
                    
                    if (text.trim()) {
                        const fileType = getFileTypeFromDocument(document);
                        const result = await this.apiClient.convertFile(text, fileType);
                        const wordCount = text.split(/\s+/).length;
                        const changeCount = result.metadata.spellingChanges + result.metadata.unitChanges;
                        
                        projectResults.push({
                            fileName: path.basename(file.fsPath),
                            relativePath: relativePath,
                            result: result,
                            wordCount: wordCount,
                            changeCount: changeCount
                        });

                        totalChanges += changeCount;
                        totalWords += wordCount;
                    }
                } catch (error) {
                    this.outputChannel.appendLine(`Error processing ${relativePath}: ${error}`);
                }
            }

            await this.showProjectReportInNewTab(projectResults, totalWords, totalChanges);
        });
    }

    /**
     * Find text files in the workspace
     */
    private async findTextFiles(_workspaceUri: vscode.Uri): Promise<vscode.Uri[]> {
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

    /**
     * Show report in a new tab
     */
    private async showReportInNewTab(result: any, fileName: string, reportType: string): Promise<void> {
        const reportContent = this.generateReportContent(result, fileName, reportType);
        
        const doc = await vscode.workspace.openTextDocument({
            content: reportContent,
            language: 'markdown'
        });

        await vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
    }

    /**
     * Show project report in a new tab
     */
    private async showProjectReportInNewTab(
        results: Array<{
            fileName: string;
            relativePath: string;
            result: any;
            wordCount: number;
            changeCount: number;
        }>,
        totalWords: number,
        totalChanges: number
    ): Promise<void> {
        const reportContent = this.generateProjectReportContent(results, totalWords, totalChanges);
        
        const doc = await vscode.workspace.openTextDocument({
            content: reportContent,
            language: 'markdown'
        });

        await vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
    }

    /**
     * Generate report content in markdown format
     */
    private generateReportContent(result: any, fileName: string, reportType: string): string {
        const changeCount = result.metadata.spellingChanges + result.metadata.unitChanges;
        const wordCount = result.originalText.split(/\s+/).length;
        
        let content = `# M2E ${reportType}: ${fileName}\n\n`;
        content += `**Generated:** ${new Date().toLocaleString()}\n\n`;
        
        content += `## Summary\n\n`;
        content += `- **Total Words:** ${wordCount}\n`;
        content += `- **Potential Changes:** ${changeCount}\n`;
        content += `- **Spelling Changes:** ${result.metadata.spellingChanges}\n`;
        content += `- **Unit Changes:** ${result.metadata.unitChanges}\n\n`;

        if (changeCount > 0) {
            content += `## Detected Changes\n\n`;
            
            if (result.changes && result.changes.length > 0) {
                for (const change of result.changes) {
                    const type = change.type === 'spelling' ? '📝 Spelling' : '📏 Unit';
                    content += `- **${type}:** \`${change.original}\` → \`${change.converted}\`\n`;
                }
            }
            
            content += `\n## Preview\n\n`;
            content += `### Original Text\n\`\`\`\n${result.originalText}\n\`\`\`\n\n`;
            content += `### Converted Text\n\`\`\`\n${result.convertedText}\n\`\`\`\n\n`;
        } else {
            content += `## Result\n\n`;
            content += `No American spellings or imperial units detected in this file.\n\n`;
        }

        content += `---\n\n`;
        content += `*Report generated by M2E VSCode Extension*\n`;

        return content;
    }

    /**
     * Generate project report content in markdown format
     */
    private generateProjectReportContent(
        results: Array<{
            fileName: string;
            relativePath: string;
            result: any;
            wordCount: number;
            changeCount: number;
        }>,
        totalWords: number,
        totalChanges: number
    ): string {
        let content = `# M2E Project Report\n\n`;
        content += `**Generated:** ${new Date().toLocaleString()}\n\n`;
        
        content += `## Project Summary\n\n`;
        content += `- **Files Analysed:** ${results.length}\n`;
        content += `- **Total Words:** ${totalWords.toLocaleString()}\n`;
        content += `- **Total Potential Changes:** ${totalChanges}\n`;
        
        const filesWithChanges = results.filter(r => r.changeCount > 0);
        content += `- **Files with Changes:** ${filesWithChanges.length}\n\n`;

        if (filesWithChanges.length > 0) {
            content += `## Files Requiring Changes\n\n`;
            
            // Sort by change count descending
            filesWithChanges.sort((a, b) => b.changeCount - a.changeCount);
            
            for (const file of filesWithChanges) {
                content += `### ${file.relativePath}\n\n`;
                content += `- **Words:** ${file.wordCount}\n`;
                content += `- **Changes:** ${file.changeCount}\n`;
                content += `- **Spelling:** ${file.result.metadata.spellingChanges}\n`;
                content += `- **Units:** ${file.result.metadata.unitChanges}\n\n`;
                
                if (file.result.changes && file.result.changes.length > 0) {
                    content += `**Changes:**\n`;
                    for (const change of file.result.changes.slice(0, 5)) { // Limit to first 5 changes
                        const type = change.type === 'spelling' ? '📝' : '📏';
                        content += `- ${type} \`${change.original}\` → \`${change.converted}\`\n`;
                    }
                    if (file.result.changes.length > 5) {
                        content += `- *...and ${file.result.changes.length - 5} more changes*\n`;
                    }
                    content += `\n`;
                }
            }
        }

        if (results.length > filesWithChanges.length) {
            content += `## Files with No Changes\n\n`;
            const filesWithoutChanges = results.filter(r => r.changeCount === 0);
            
            for (const file of filesWithoutChanges) {
                content += `- ${file.relativePath} (${file.wordCount} words)\n`;
            }
            content += `\n`;
        }

        content += `---\n\n`;
        content += `*Project report generated by M2E VSCode Extension*\n`;

        return content;
    }
}