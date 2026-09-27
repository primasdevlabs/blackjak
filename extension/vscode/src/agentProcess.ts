import * as child_process from 'child_process';
import * as path from 'path';
import * as fs from 'fs';
import * as vscode from 'vscode';

export interface AgentProcessInfo {
  process: child_process.ChildProcess;
  port: number;
  workspacePath: string;
}

export class AgentProcessManager {
  private childProcess: child_process.ChildProcess | null = null;
  private port: number = 0;
  private outputChannel: vscode.OutputChannel;

  constructor(outputChannel: vscode.OutputChannel) {
    this.outputChannel = outputChannel;
  }

  public async startProcess(workspacePath: string, extensionPath: string): Promise<{ process: child_process.ChildProcess; port: number }> {
    if (this.childProcess) {
      await this.stopProcess();
    }

    const binaryPath = this.resolveBinaryPath(extensionPath);
    this.outputChannel.appendLine(`[AgentProcess] Launching binary: ${binaryPath}`);
    this.outputChannel.appendLine(`[AgentProcess] Workspace: ${workspacePath}`);

    return new Promise((resolve, reject) => {
      const args = ['--server', '--port=0', `--workspace=${workspacePath}`];
      let spawned: child_process.ChildProcess;

      if (binaryPath.endsWith('.go')) {
        // Fallback for development if Go source file path detected
        spawned = child_process.spawn('go', ['run', binaryPath, ...args], {
          cwd: path.dirname(binaryPath),
          env: process.env,
        });
      } else {
        spawned = child_process.spawn(binaryPath, args, {
          cwd: workspacePath,
          env: process.env,
        });
      }

      this.childProcess = spawned;
      let detectedPort = 0;
      let startupResolved = false;

      const timeoutTimer = setTimeout(() => {
        if (!startupResolved) {
          startupResolved = true;
          if (detectedPort > 0) {
            resolve({ process: spawned, port: detectedPort });
          } else {
            reject(new Error('Agent process startup timed out waiting for port allocation'));
          }
        }
      }, 10000);

      spawned.stdout?.on('data', (data: Buffer) => {
        const text = data.toString();
        this.outputChannel.appendLine(`[Agent stdout] ${text.trim()}`);

        // Parse JSON startup line: {"status":"ok","ready":true,...,"port":48123}
        const jsonMatch = text.match(/\{"status":"ok","ready":true.*"port":(\d+).*\}/);
        if (jsonMatch && jsonMatch[1]) {
          detectedPort = parseInt(jsonMatch[1], 10);
          this.port = detectedPort;
          if (!startupResolved) {
            startupResolved = true;
            clearTimeout(timeoutTimer);
            resolve({ process: spawned, port: detectedPort });
          }
        }
      });

      spawned.stderr?.on('data', (data: Buffer) => {
        this.outputChannel.appendLine(`[Agent stderr] ${data.toString().trim()}`);
      });

      spawned.on('error', (err) => {
        this.outputChannel.appendLine(`[AgentProcess Error] ${err.message}`);
        if (!startupResolved) {
          startupResolved = true;
          clearTimeout(timeoutTimer);
          reject(err);
        }
      });

      spawned.on('exit', (code, signal) => {
        this.outputChannel.appendLine(`[AgentProcess Exit] Process terminated with code ${code}, signal ${signal}`);
        this.childProcess = null;
        this.port = 0;
      });
    });
  }

  public async stopProcess(): Promise<void> {
    if (!this.childProcess) return;

    this.outputChannel.appendLine('[AgentProcess] Terminating Go backend process...');
    return new Promise((resolve) => {
      if (!this.childProcess) {
        resolve();
        return;
      }
      const proc = this.childProcess;
      this.childProcess = null;

      proc.once('exit', () => {
        this.outputChannel.appendLine('[AgentProcess] Stopped cleanly.');
        resolve();
      });

      proc.kill('SIGTERM');
      setTimeout(() => {
        try {
          proc.kill('SIGKILL');
        } catch (_) {}
        resolve();
      }, 3000);
    });
  }

  public getPort(): number {
    return this.port;
  }

  private resolveBinaryPath(extensionPath: string): string {
    const platform = process.platform;
    const arch = process.arch;

    let binaryName = `agent-${platform}-${arch}`;
    if (platform === 'win32') {
      binaryName += '.exe';
    }

    // 1. Check extension bundled bin/ directory
    const bundledPath = path.join(extensionPath, 'bin', binaryName);
    if (fs.existsSync(bundledPath)) {
      return bundledPath;
    }

    // 2. Check workspace root bin/ binary
    const workspaceBinary = path.join(extensionPath, '..', '..', 'bin', platform === 'win32' ? 'agent.exe' : 'agent');
    if (fs.existsSync(workspaceBinary)) {
      return workspaceBinary;
    }

    // 3. Fallback to cmd/agent/main.go for development
    const goSourcePath = path.join(extensionPath, '..', '..', 'cmd', 'agent', 'main.go');
    if (fs.existsSync(goSourcePath)) {
      return goSourcePath;
    }

    return bundledPath;
  }
}
