import * as http from 'http';
import * as url from 'url';

/**
 * Mock M2E server for testing purposes
 */
export class MockM2EServer {
    private server: http.Server;
    private port: number;
    private responses: Map<string, any> = new Map();

    constructor(port: number = 18182) { // Use different port from real server
        this.port = port;
        this.server = http.createServer(this.handleRequest.bind(this));
        this.setupDefaultResponses();
    }

    /**
     * Start the mock server
     */
    async start(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.server.listen(this.port, (err?: Error) => {
                if (err) {
                    reject(err);
                } else {
                    console.log(`Mock M2E server running on port ${this.port}`);
                    resolve();
                }
            });
        });
    }

    /**
     * Stop the mock server
     */
    async stop(): Promise<void> {
        return new Promise((resolve) => {
            this.server.close(() => {
                console.log('Mock M2E server stopped');
                resolve();
            });
        });
    }

    /**
     * Get server URL
     */
    getUrl(): string {
        return `http://localhost:${this.port}`;
    }

    /**
     * Set custom response for a specific endpoint
     */
    setResponse(endpoint: string, response: any): void {
        this.responses.set(endpoint, response);
    }

    /**
     * Setup default responses for common endpoints
     */
    private setupDefaultResponses(): void {
        // Health check endpoint
        this.responses.set('/api/v1/health', {
            status: 'healthy',
            version: '1.0.0',
            timestamp: new Date().toISOString()
        });

        // Convert text endpoint
        this.responses.set('/api/v1/convert', {
            originalText: 'color',
            convertedText: 'colour',
            metadata: {
                spellingChanges: 1,
                unitChanges: 0,
                processingTimeMs: 10,
                fileType: 'text'
            }
        });

        // Convert file endpoint  
        this.responses.set('/api/v1/convert-file', {
            originalText: 'The color is great',
            convertedText: 'The colour is great',
            metadata: {
                spellingChanges: 1,
                unitChanges: 0,
                processingTimeMs: 15,
                fileType: 'text'
            }
        });

        // Convert comments endpoint
        this.responses.set('/api/v1/convert-comments', {
            originalText: '// The color is great\nfunction test() {}',
            convertedText: '// The colour is great\nfunction test() {}',
            metadata: {
                spellingChanges: 1,
                unitChanges: 0,
                processingTimeMs: 12,
                fileType: 'javascript'
            }
        });
    }

    /**
     * Handle incoming HTTP requests
     */
    private handleRequest(req: http.IncomingMessage, res: http.ServerResponse): void {
        const parsedUrl = url.parse(req.url || '', true);
        const pathname = parsedUrl.pathname || '';

        // Set CORS headers
        res.setHeader('Access-Control-Allow-Origin', '*');
        res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
        res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

        // Handle OPTIONS requests
        if (req.method === 'OPTIONS') {
            res.writeHead(200);
            res.end();
            return;
        }

        // Handle GET requests
        if (req.method === 'GET') {
            const response = this.responses.get(pathname);
            if (response) {
                res.writeHead(200, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify(response));
            } else {
                res.writeHead(404, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify({ error: 'Not found' }));
            }
            return;
        }

        // Handle POST requests
        if (req.method === 'POST') {
            let body = '';
            req.on('data', chunk => {
                body += chunk.toString();
            });

            req.on('end', () => {
                try {
                    const requestData = JSON.parse(body);
                    const response = this.handlePostRequest(pathname, requestData);
                    
                    res.writeHead(200, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify(response));
                } catch (error) {
                    res.writeHead(400, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify({ error: 'Invalid JSON' }));
                }
            });
            return;
        }

        // Method not allowed
        res.writeHead(405, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: 'Method not allowed' }));
    }

    /**
     * Handle POST requests with dynamic responses
     */
    private handlePostRequest(pathname: string, requestData: any): any {
        switch (pathname) {
            case '/api/v1/convert':
                return this.handleConvertRequest(requestData);
            
            case '/api/v1/convert-file':
                return this.handleConvertFileRequest(requestData);
            
            case '/api/v1/convert-comments':
                return this.handleConvertCommentsRequest(requestData);
            
            default:
                return { error: 'Endpoint not found' };
        }
    }

    /**
     * Handle convert text requests
     */
    private handleConvertRequest(data: any): any {
        const text = data.text || '';
        const changes: Array<{
            position: number;
            original: string;
            converted: string;
            type: string;
            is_contextual?: boolean;
        }> = [];
        
        // Define spelling conversions
        const spellingMap = new Map<string, string>([
            ['color', 'colour'],
            ['colors', 'colours'],
            ['organize', 'organise'],
            ['organizes', 'organises'],
            ['organization', 'organisation'],
            ['analyze', 'analyse'],
            ['center', 'centre'],
            ['realize', 'realise']
        ]);
        
        // Find all American spellings first (before any conversions that might shift positions)
        for (const [american, british] of spellingMap.entries()) {
            const regex = new RegExp(`\\b${american}\\b`, 'gi');
            let match;
            
            while ((match = regex.exec(text)) !== null) {
                const originalWord = match[0];
                const convertedWord = this.preserveCase(originalWord, british);
                
                changes.push({
                    position: match.index,
                    original: originalWord,
                    converted: convertedWord,
                    type: 'spelling',
                    is_contextual: false
                });
            }
        }
        
        // Sort changes by position (descending) to avoid position shifts when applying
        changes.sort((a, b) => b.position - a.position);
        
        // Apply conversions from right to left to avoid position shifts
        let convertedText = text;
        for (const change of changes) {
            convertedText = convertedText.substring(0, change.position) + 
                          change.converted + 
                          convertedText.substring(change.position + change.original.length);
        }
        
        // Sort changes back by position (ascending) for the response
        changes.sort((a, b) => a.position - b.position);

        return {
            text: convertedText,
            changes: changes
        };
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

    /**
     * Handle convert file requests
     */
    private handleConvertFileRequest(data: any): any {
        return this.handleConvertRequest(data);
    }

    /**
     * Handle convert comments requests
     */
    private handleConvertCommentsRequest(data: any): any {
        return this.handleConvertRequest(data); // Use same logic for now
    }
}