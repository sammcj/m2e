"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.MockM2EServer = void 0;
const http = __importStar(require("http"));
const url = __importStar(require("url"));
/**
 * Mock M2E server for testing purposes
 */
class MockM2EServer {
    constructor(port = 18182) {
        this.responses = new Map();
        this.port = port;
        this.server = http.createServer(this.handleRequest.bind(this));
        this.setupDefaultResponses();
    }
    /**
     * Start the mock server
     */
    async start() {
        return new Promise((resolve, reject) => {
            this.server.listen(this.port, (err) => {
                if (err) {
                    reject(err);
                }
                else {
                    console.log(`Mock M2E server running on port ${this.port}`);
                    resolve();
                }
            });
        });
    }
    /**
     * Stop the mock server
     */
    async stop() {
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
    getUrl() {
        return `http://localhost:${this.port}`;
    }
    /**
     * Set custom response for a specific endpoint
     */
    setResponse(endpoint, response) {
        this.responses.set(endpoint, response);
    }
    /**
     * Setup default responses for common endpoints
     */
    setupDefaultResponses() {
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
    handleRequest(req, res) {
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
            }
            else {
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
                }
                catch (error) {
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
    handlePostRequest(pathname, requestData) {
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
    handleConvertRequest(data) {
        const text = data.text || '';
        const convertedText = text.replace(/color/g, 'colour').replace(/organize/g, 'organise');
        const changes = (text.match(/color|organize/g) || []).length;
        return {
            originalText: text,
            convertedText: convertedText,
            metadata: {
                spellingChanges: changes,
                unitChanges: 0,
                processingTimeMs: Math.floor(Math.random() * 20) + 5,
                fileType: data.fileType || 'text'
            }
        };
    }
    /**
     * Handle convert file requests
     */
    handleConvertFileRequest(data) {
        return this.handleConvertRequest(data);
    }
    /**
     * Handle convert comments requests
     */
    handleConvertCommentsRequest(data) {
        const text = data.text || '';
        // Only convert in comments (lines starting with //)
        const lines = text.split('\n');
        const convertedLines = lines.map(line => {
            if (line.trim().startsWith('//')) {
                return line.replace(/color/g, 'colour').replace(/organize/g, 'organise');
            }
            return line;
        });
        const convertedText = convertedLines.join('\n');
        const changes = text.match(/(\/\/.*?color|\/\/.*?organize)/g)?.length || 0;
        return {
            originalText: text,
            convertedText: convertedText,
            metadata: {
                spellingChanges: changes,
                unitChanges: 0,
                processingTimeMs: Math.floor(Math.random() * 20) + 5,
                fileType: data.fileType || 'javascript'
            }
        };
    }
}
exports.MockM2EServer = MockM2EServer;
//# sourceMappingURL=mockServer.js.map