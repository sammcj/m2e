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
     * Generate changes array by finding specific word replacements in original text
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

        // Define common American to British spelling conversions to look for
        const spellingMap = new Map<string, string>([
            ['color', 'colour'],
            ['colors', 'colours'],
            ['colored', 'coloured'],
            ['coloring', 'colouring'],
            ['colorful', 'colourful'],
            ['organize', 'organise'],
            ['organizes', 'organises'], 
            ['organized', 'organised'],
            ['organizing', 'organising'],
            ['organization', 'organisation'],
            ['organizations', 'organisations'],
            ['center', 'centre'],
            ['centers', 'centres'],
            ['centered', 'centred'],
            ['centering', 'centring'],
            ['analyze', 'analyse'],
            ['analyzes', 'analyses'],
            ['analyzed', 'analysed'],
            ['analyzing', 'analysing'],
            ['analyzer', 'analyser'],
            ['realize', 'realise'],
            ['realizes', 'realises'],
            ['realized', 'realised'],
            ['realizing', 'realising'],
            ['realization', 'realisation'],
            ['aluminum', 'aluminium'],
            ['honor', 'honour'],
            ['honors', 'honours'],
            ['honored', 'honoured'],
            ['honoring', 'honouring'],
            ['flavor', 'flavour'],
            ['flavors', 'flavours'],
            ['flavored', 'flavoured'],
            ['flavoring', 'flavouring'],
            ['neighbor', 'neighbour'],
            ['neighbors', 'neighbours'],
            ['neighborhood', 'neighbourhood'],
            ['defense', 'defence'],
            ['offense', 'offence']
        ]);

        // Look for each American spelling in the original text
        for (const [american, british] of spellingMap.entries()) {
            // Create regex to find whole words only, case insensitive
            const regex = new RegExp(`\\b${american}\\b`, 'gi');
            let match;
            
            while ((match = regex.exec(originalText)) !== null) {
                // Check if this word was actually converted in the result
                const originalWord = match[0];
                const expectedBritish = this.preserveCase(originalWord, british);
                
                // Verify the conversion actually happened by checking if the British version exists in converted text
                if (convertedText.includes(expectedBritish)) {
                    changes.push({
                        position: match.index,
                        original: originalWord,
                        converted: expectedBritish,
                        type: 'spelling'
                    });
                }
            }
        }

        // Look for unit conversions (simple patterns)
        const unitPatterns = [
            // Temperature
            { pattern: /(\d+(?:\.\d+)?)\s*°F\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*degrees?\s+fahrenheit/gi, type: 'unit' as const },
            // Distance 
            { pattern: /(\d+(?:\.\d+)?)\s*feet?\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*ft\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*inches?\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*in\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*miles?\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*yards?\b/g, type: 'unit' as const },
            // Weight
            { pattern: /(\d+(?:\.\d+)?)\s*pounds?\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*lbs?\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*ounces?\b/g, type: 'unit' as const },
            { pattern: /(\d+(?:\.\d+)?)\s*oz\b/g, type: 'unit' as const }
        ];

        for (const unitPattern of unitPatterns) {
            let match;
            while ((match = unitPattern.pattern.exec(originalText)) !== null) {
                // Only add if this unit appears to be converted (length changed significantly)
                const originalLength = originalText.length;
                const convertedLength = convertedText.length;
                
                if (Math.abs(convertedLength - originalLength) > 0) {
                    changes.push({
                        position: match.index,
                        original: match[0],
                        converted: match[0], // We don't know the exact conversion without more analysis
                        type: 'unit'
                    });
                }
            }
        }

        return changes;
    }

    /**
     * Preserve the original case pattern when converting words
     */
    private preserveCase(original: string, converted: string): string {
        if (original === original.toUpperCase()) {
            return converted.toUpperCase();
        } else if (original === original.toLowerCase()) {
            return converted.toLowerCase();
        } else if (original[0] === original[0].toUpperCase()) {
            return converted.charAt(0).toUpperCase() + converted.slice(1).toLowerCase();
        }
        return converted;
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