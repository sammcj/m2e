import * as vscode from 'vscode';
import { spawn, ChildProcess } from 'child_process';
import * as path from 'path';
import * as fs from 'fs';
import * as net from 'net';

export interface ServerStatus {
    isRunning: boolean;
    port: number;
    pid?: number;
    error?: string;
}

export class M2EServerManager {
    private process: ChildProcess | null = null;
    private port: number = 18181;
    private outputChannel: vscode.OutputChannel;
    private context: vscode.ExtensionContext;
    private healthCheckTimer?: ReturnType<typeof setInterval> | undefined;

    constructor(
        context: vscode.ExtensionContext,
        outputChannel: vscode.OutputChannel
    ) {
        this.context = context;
        this.outputChannel = outputChannel;
        
        // Register cleanup handlers
        context.subscriptions.push(
            vscode.workspace.onDidChangeConfiguration(this.onConfigChanged.bind(this)),
            { dispose: () => this.stop() }
        );
    }

    async start(): Promise<boolean> {
        try {
            // Check if server is already running
            if (this.process && !this.process.killed) {
                this.outputChannel.appendLine('M2E server is already running');
                return true;
            }

            // Get binary path based on platform
            const binaryPath = this.getBinaryPath();
            if (!binaryPath) {
                return false;
            }

            // Find available port
            this.port = await this.findAvailablePort();

            // Start server process
            this.outputChannel.appendLine(`Starting M2E server on port ${this.port}`);
            this.process = spawn(binaryPath, [], {
                env: { 
                    ...process.env, 
                    API_PORT: this.port.toString(),
                    M2E_LOG_LEVEL: this.getLogLevel()
                },
                stdio: ['ignore', 'pipe', 'pipe']
            });

            // Handle server output
            this.process.stdout?.on('data', (data) => {
                this.outputChannel.appendLine(`[Server] ${data.toString().trim()}`);
            });

            this.process.stderr?.on('data', (data) => {
                this.outputChannel.appendLine(`[Server Error] ${data.toString().trim()}`);
            });

            // Handle process events
            this.process.on('error', (_error) => {
                this.outputChannel.appendLine(`[Server Process Error] ${_error.message}`);
            });

            this.process.on('exit', (code, signal) => {
                this.outputChannel.appendLine(`[Server] Process exited with code ${code}, signal ${signal}`);
                this.process = null;
                
                // Clear health check timer
                if (this.healthCheckTimer) {
                    clearInterval(this.healthCheckTimer);
                    this.healthCheckTimer = undefined;
                }
            });

            // Wait for server to be ready
            const ready = await this.waitForServer(5000); // 5 second timeout

            if (ready) {
                this.outputChannel.appendLine(`M2E server started successfully on port ${this.port}`);
                
                // Start periodic health checks
                this.startHealthChecks();
                return true;
            } else {
                this.outputChannel.appendLine('M2E server failed to start within timeout');
                await this.stop();
                return false;
            }
        } catch {
            const message = "An unknown error occurred";
            this.outputChannel.appendLine(`Failed to start M2E server: ${message}`);
            return false;
        }
    }

    async stop(): Promise<void> {
        if (this.healthCheckTimer) {
            clearInterval(this.healthCheckTimer);
            this.healthCheckTimer = undefined;
        }

        if (this.process && !this.process.killed) {
            this.outputChannel.appendLine('Stopping M2E server...');
            
            // Try graceful shutdown first
            this.process.kill('SIGTERM');
            
            // Wait for graceful shutdown, then force kill if needed
            await new Promise<void>((resolve) => {
                const timeout = setTimeout(() => {
                    if (this.process && !this.process.killed) {
                        this.outputChannel.appendLine('Force killing M2E server process');
                        this.process.kill('SIGKILL');
                    }
                    resolve();
                }, 3000); // 3 second timeout for graceful shutdown

                if (this.process) {
                    this.process.on('exit', () => {
                        clearTimeout(timeout);
                        resolve();
                    });
                }
            });

            this.process = null;
            this.outputChannel.appendLine('M2E server stopped');
        }
    }

    async restart(): Promise<boolean> {
        this.outputChannel.appendLine('Restarting M2E server...');
        await this.stop();
        return await this.start();
    }

    getStatus(): ServerStatus {
        const status: ServerStatus = {
            isRunning: this.process !== null && !this.process.killed,
            port: this.port
        };
        
        if (this.process?.pid !== undefined) {
            status.pid = this.process.pid;
        }
        
        if (this.process?.killed) {
            status.error = 'Process was killed';
        }
        
        return status;
    }

    getServerUrl(): string {
        return `http://localhost:${this.port}`;
    }

    private getBinaryPath(): string | null {
        const platform = process.platform;
        const arch = process.arch;

        // Check if running on Windows (not supported)
        if (platform === 'win32') {
            vscode.window.showErrorMessage(
                'M2E: Windows is not supported. Only macOS and Linux are supported.'
            );
            return null;
        }

        // Map platform and architecture to our binary directory structure
        let binaryDir: string;
        if (platform === 'darwin') {
            if (arch === 'arm64') {
                binaryDir = 'darwin-arm64';
            } else {
                // For Intel Macs, use the arm64 binary (works via Rosetta)
                binaryDir = 'darwin-arm64';
            }
        } else if (platform === 'linux') {
            binaryDir = arch === 'arm64' ? 'linux-arm64' : 'linux-x64';
        } else {
            vscode.window.showErrorMessage(
                `M2E: Platform ${platform} is not supported. Only macOS and Linux are supported.`
            );
            return null;
        }

        // Check for custom path first
        const customPath = vscode.workspace.getConfiguration('m2e').get<string>('customServerPath');
        if (customPath && fs.existsSync(customPath)) {
            this.outputChannel.appendLine(`Using custom server binary: ${customPath}`);
            return customPath;
        }

        // Use bundled binary
        const binaryPath = path.join(
            this.context.extensionPath,
            'resources',
            'bin',
            binaryDir,
            'm2e-server'
        );

        if (!fs.existsSync(binaryPath)) {
            const message = `M2E server binary not found at ${binaryPath}. Please reinstall the extension.`;
            vscode.window.showErrorMessage(message);
            this.outputChannel.appendLine(message);
            return null;
        }

        try {
            // Ensure binary is executable
            fs.chmodSync(binaryPath, 0o755);
            this.outputChannel.appendLine(`Using bundled server binary: ${binaryPath}`);
            return binaryPath;
        } catch {
            const message = `Failed to make server binary executable`;
            vscode.window.showErrorMessage(message);
            this.outputChannel.appendLine(message);
            return null;
        }
    }

    private async findAvailablePort(): Promise<number> {
        const startPort = vscode.workspace.getConfiguration('m2e').get<number>('serverPort') || 18181;

        for (let port = startPort; port < startPort + 20; port++) {
            if (await this.isPortAvailable(port)) {
                return port;
            }
        }

        throw new Error(`No available ports found in range ${startPort}-${startPort + 19}`);
    }

    private async isPortAvailable(port: number): Promise<boolean> {
        return new Promise<boolean>((resolve) => {
            const server = net.createServer();
            
            server.listen(port, () => {
                server.close(() => resolve(true));
            });
            
            server.on('error', () => resolve(false));
        });
    }

    private async waitForServer(timeout: number = 5000): Promise<boolean> {
        const startTime = Date.now();
        const interval = 100; // Check every 100ms

        while (Date.now() - startTime < timeout) {
            try {
                // Use node's net module for faster health check
                const isHealthy = await this.checkHealth();
                if (isHealthy) {
                    return true;
                }
            } catch {
                // Server not ready yet
            }

            await new Promise(resolve => setTimeout(resolve, interval));
        }

        return false;
    }

    private async checkHealth(): Promise<boolean> {
        return new Promise<boolean>((resolve) => {
            const socket = new net.Socket();
            const timeout = 1000; // 1 second timeout for health checks

            const timer = setTimeout(() => {
                socket.destroy();
                resolve(false);
            }, timeout);

            socket.connect(this.port, 'localhost', () => {
                clearTimeout(timer);
                socket.destroy();
                resolve(true);
            });

            socket.on('error', () => {
                clearTimeout(timer);
                resolve(false);
            });
        });
    }

    private startHealthChecks(): void {
        // Perform health check every 30 seconds
        this.healthCheckTimer = setInterval(async () => {
            if (this.process && !this.process.killed) {
                const isHealthy = await this.checkHealth();
                if (!isHealthy) {
                    this.outputChannel.appendLine('Health check failed - server may be unresponsive');
                }
            }
        }, 30000);
    }


    private getLogLevel(): string {
        const debugLogging = vscode.workspace.getConfiguration('m2e').get<boolean>('debugLogging', false);
        return debugLogging ? 'debug' : 'warn';
    }

    private async onConfigChanged(event: vscode.ConfigurationChangeEvent): Promise<void> {
        if (event.affectsConfiguration('m2e.serverPort') || 
            event.affectsConfiguration('m2e.customServerPath') ||
            event.affectsConfiguration('m2e.debugLogging')) {
            
            this.outputChannel.appendLine('M2E configuration changed, restarting server...');
            await this.restart();
        }
    }
}