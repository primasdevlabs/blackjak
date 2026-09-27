import * as vscode from 'vscode';
import { spawn, ChildProcess } from 'child_process';
import * as http from 'http';
import * as path from 'path';

export class AgentServerManager {
  private process: ChildProcess | null = null;
  private port: number = 8080;
  private host: string = '127.0.0.1';
  private workspacePath: string;

  constructor(workspacePath: string, port: number = 8080) {
    this.workspacePath = workspacePath;
    this.port = port;
  }

  getPort(): number {
    return this.port;
  }

  getHost(): string {
    return this.host;
  }

  async ensureServerRunning(extensionPath: string): Promise<boolean> {
    const isHealthy = await this.checkHealth();
    if (isHealthy) {
      return true;
    }

    return this.startServer(extensionPath);
  }

  async checkHealth(): Promise<boolean> {
    return new Promise((resolve) => {
      const req = http.get(`http://${this.host}:${this.port}/health`, (res) => {
        if (res.statusCode === 200) {
          resolve(true);
        } else {
          resolve(false);
        }
      });

      req.on('error', () => {
        resolve(false);
      });

      req.setTimeout(1000, () => {
        req.destroy();
        resolve(false);
      });
    });
  }

  private async startServer(extensionPath: string): Promise<boolean> {
    const projectRoot = path.resolve(extensionPath, '..', '..');

    vscode.window.showInformationMessage('Starting Go Agent backend server...');

    try {
      this.process = spawn('go', [
        'run',
        'cmd/agent/main.go',
        '--server',
        '--host',
        this.host,
        '--port',
        this.port.toString(),
        '--workspace',
        this.workspacePath,
      ], {
        cwd: projectRoot,
        env: { ...process.env },
      });

      this.process.stdout?.on('data', (data) => {
        console.log(`[Agent Go Server]: ${data}`);
      });

      this.process.stderr?.on('data', (data) => {
        console.error(`[Agent Go Server Error]: ${data}`);
      });

      this.process.on('exit', (code) => {
        console.log(`[Agent Go Server] Exited with code ${code}`);
        this.process = null;
      });

      // Poll health check for up to 10 seconds
      for (let i = 0; i < 20; i++) {
        await new Promise((r) => setTimeout(r, 500));
        if (await this.checkHealth()) {
          vscode.window.showInformationMessage(`Agent server ready on port ${this.port}`);
          return true;
        }
      }

      vscode.window.showErrorMessage('Agent server failed to start within timeout.');
      return false;
    } catch (err: any) {
      vscode.window.showErrorMessage(`Failed to spawn Go agent process: ${err.message}`);
      return false;
    }
  }

  stopServer() {
    if (this.process) {
      this.process.kill();
      this.process = null;
    }
  }
}
