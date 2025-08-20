import fetch, { RequestInit, Response } from 'node-fetch';
import * as vscode from 'vscode';

export interface ConvertRequest {
    text: string;
    options?: {
        enableUnitConversion?: boolean;
        codeAware?: boolean;
        preserveCodeSyntax?: boolean;
        fileType?: string;
        commentsOnly?: boolean;
    };
}

export interface ConvertResponse {
    originalText: string;
    convertedText: string;
    changes: Array<{
        position: number;
        original: string;
        converted: string;
        type: 'spelling' | 'unit';
        is_contextual?: boolean;
    }>;
    metadata: {
        spellingChanges: number;
        unitChanges: number;
        processingTimeMs: number;
    };
}

// Internal API response from server
interface ApiResponse {
    text: string;
    changes?: Array<{
        position: number;
        original: string;
        converted: string;
        type: string;
        is_contextual?: boolean;
    }>;
}

export interface HealthResponse {
    status: 'healthy';
    version: string;
    uptime: number;
}

export class M2EApiClient {
    private baseUrl: string = '';
    private outputChannel: vscode.OutputChannel;

    constructor(outputChannel: vscode.OutputChannel) {
        this.outputChannel = outputChannel;
    }

    setServerUrl(url: string): void {
        this.baseUrl = url;
    }

    async convert(request: ConvertRequest): Promise<ConvertResponse> {
        try {
            // Merge request options with user configuration
            const config = vscode.workspace.getConfiguration('m2e');
            const options = {
                enableUnitConversion: config.get<boolean>('enableUnitConversion', true),
                codeAware: config.get<boolean>('codeAwareConversion', true),
                preserveCodeSyntax: config.get<boolean>('preserveCodeSyntax', true),
                ...request.options
            };

            this.logDebug(`Converting text: ${request.text.substring(0, 100)}${request.text.length > 100 ? '...' : ''}`);

            const response = await this.makeRequest('/api/v1/convert', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    text: request.text,
                    convert_units: options.enableUnitConversion,
                    normalise_smart_quotes: true
                })
            });

            const apiResult = await response.json() as ApiResponse;
            
            // Transform API response to match our interface
            let changes: ConvertResponse['changes'] = [];
            
            if (apiResult.changes && apiResult.changes.length > 0) {
                // Use changes from the enhanced API response
                changes = apiResult.changes.map(change => ({
                    position: change.position,
                    original: change.original,
                    converted: change.converted,
                    type: change.type as 'spelling' | 'unit',
                    ...(change.is_contextual !== undefined && { is_contextual: change.is_contextual })
                }));
            } else {
                // Fallback to generating changes if not provided
                changes = this.generateChanges(request.text, apiResult.text);
            }
            
            const spellingChanges = changes.filter(c => c.type === 'spelling').length;
            const unitChanges = changes.filter(c => c.type === 'unit').length;
            
            const result: ConvertResponse = {
                originalText: request.text,
                convertedText: apiResult.text,
                changes: changes,
                metadata: {
                    spellingChanges: spellingChanges,
                    unitChanges: unitChanges,
                    processingTimeMs: 0  // Not available from simple API
                }
            };
            
            this.logDebug(`Conversion completed: ${result.metadata.spellingChanges} spelling changes, ${result.metadata.unitChanges} unit changes`);
            
            return result;
        } catch {
            const message = "An unknown error occurred";
            this.outputChannel.appendLine(`[API] Convert request failed: ${message}`);
            throw new Error(`Failed to convert text: ${message}`);
        }
    }

    async checkHealth(): Promise<HealthResponse> {
        try {
            const response = await this.makeRequest('/api/v1/health', {
                method: 'GET'
            });

            // The actual API just returns "OK", so we create our own response
            await response.text(); // Read the "OK" response
            
            const result: HealthResponse = {
                status: 'healthy',
                version: '1.0.0',
                uptime: 0
            };
            
            this.logDebug(`Health check successful`);
            
            return result;
        } catch {
            const message = "An unknown error occurred";
            this.outputChannel.appendLine(`[API] Health check failed: ${message}`);
            throw new Error(`Health check failed: ${message}`);
        }
    }

    async convertSelection(text: string, fileType?: string): Promise<ConvertResponse> {
        return this.convert({
            text,
            options: fileType ? {
                fileType,
                commentsOnly: false
            } : {
                commentsOnly: false
            }
        });
    }

    async convertFile(text: string, fileType?: string): Promise<ConvertResponse> {
        return this.convert({
            text,
            options: fileType ? {
                fileType,
                commentsOnly: false
            } : {
                commentsOnly: false
            }
        });
    }

    async convertCommentsOnly(text: string, fileType?: string): Promise<ConvertResponse> {
        if (!fileType) {
            throw new Error('File type is required for comments-only conversion');
        }

        return this.convert({
            text,
            options: {
                fileType,
                commentsOnly: true
            }
        });
    }

    private async makeRequest(endpoint: string, options: RequestInit): Promise<Response> {
        if (!this.baseUrl) {
            throw new Error('Server URL not set. Make sure the M2E server is running.');
        }

        const url = `${this.baseUrl}${endpoint}`;
        const timeout = 30000; // 30 second timeout for API requests

        // Create abort controller for timeout
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), timeout);

        try {
            const response = await fetch(url, {
                ...options,
                signal: controller.signal,
                // Add user agent
                headers: {
                    'User-Agent': 'M2E-VSCode-Extension',
                    ...options.headers
                }
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
                
                try {
                    const errorBody = await response.text();
                    if (errorBody) {
                        errorMessage += ` - ${errorBody}`;
                    }
                } catch {
                    // Ignore errors reading response body
                }

                throw new Error(errorMessage);
            }

            return response;
        } catch {
            clearTimeout(timeoutId);
            throw new Error('Cannot connect to M2E server. Please check if the server is running.');
        }
    }

    private logDebug(message: string): void {
        const debugLogging = vscode.workspace.getConfiguration('m2e').get<boolean>('debugLogging', false);
        if (debugLogging) {
            this.outputChannel.appendLine(`[API Debug] ${message}`);
        }
    }

    /**
     * Generate changes array by comparing original and converted text
     */
    private generateChanges(originalText: string, convertedText: string): Array<{
        position: number;
        original: string;
        converted: string;
        type: 'spelling' | 'unit';
    }> {
        const changes: Array<{
            position: number;
            original: string;
            converted: string;
            type: 'spelling' | 'unit';
        }> = [];

        if (originalText === convertedText) {
            return changes;
        }

        // Simple word-by-word comparison to find changes
        const originalWords = originalText.split(/(\s+)/);
        const convertedWords = convertedText.split(/(\s+)/);
        
        let position = 0;
        const maxLength = Math.max(originalWords.length, convertedWords.length);
        
        for (let i = 0; i < maxLength; i++) {
            const originalWord = originalWords[i] || '';
            const convertedWord = convertedWords[i] || '';
            
            if (originalWord !== convertedWord && originalWord.trim() && convertedWord.trim()) {
                // Simple heuristic: if contains numbers, likely unit conversion
                const isUnitChange = /\d/.test(originalWord) || /\d/.test(convertedWord);
                
                changes.push({
                    position: position,
                    original: originalWord,
                    converted: convertedWord,
                    type: isUnitChange ? 'unit' : 'spelling'
                });
            }
            
            position += originalWord.length;
        }

        return changes;
    }
}

/**
 * Helper function to get file type from document language ID or file extension
 */
export function getFileTypeFromDocument(document: vscode.TextDocument): string | undefined {
    // Map VSCode language IDs to file types that the M2E server understands
    const languageMap: { [key: string]: string } = {
        'typescript': 'typescript',
        'javascript': 'javascript',
        'python': 'python',
        'go': 'go',
        'java': 'java',
        'c': 'c',
        'cpp': 'cpp',
        'csharp': 'csharp',
        'php': 'php',
        'ruby': 'ruby',
        'rust': 'rust',
        'markdown': 'markdown',
        'plaintext': 'text',
        'text': 'text'
    };

    // Try language ID first
    if (document.languageId && languageMap[document.languageId]) {
        return languageMap[document.languageId];
    }

    // Fall back to file extension
    const fileName = document.fileName;
    if (fileName) {
        const extension = fileName.split('.').pop()?.toLowerCase();
        if (extension) {
            const extensionMap: { [key: string]: string } = {
                'ts': 'typescript',
                'js': 'javascript',
                'py': 'python',
                'go': 'go',
                'java': 'java',
                'c': 'c',
                'cpp': 'cpp',
                'cc': 'cpp',
                'cxx': 'cpp',
                'cs': 'csharp',
                'php': 'php',
                'rb': 'ruby',
                'rs': 'rust',
                'md': 'markdown',
                'txt': 'text'
            };
            
            return extensionMap[extension];
        }
    }

    // Default to text if unknown
    return 'text';
}

/**
 * Helper function to create diff ranges for highlighting changes
 */
export function createDiffRanges(changes: ConvertResponse['changes'], document: vscode.TextDocument): vscode.Range[] {
    const ranges: vscode.Range[] = [];
    
    for (const change of changes) {
        try {
            const position = document.positionAt(change.position);
            const endPosition = document.positionAt(change.position + change.original.length);
            ranges.push(new vscode.Range(position, endPosition));
        } catch {
            // Skip invalid positions
            console.warn('Invalid position in change:', change);
        }
    }
    
    return ranges;
}