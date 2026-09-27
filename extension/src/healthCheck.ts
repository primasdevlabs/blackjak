import * as http from 'http';

export interface HealthStatus {
  status: string;
  ready: boolean;
  version: string;
  protocolVersion: string;
  port: number;
  workspace: string;
}

export class HealthCheckProbe {
  public static async pollUntilReady(port: number, timeoutMs: number = 15000): Promise<HealthStatus> {
    const startTime = Date.now();
    let delay = 200;

    while (Date.now() - startTime < timeoutMs) {
      try {
        const health = await HealthCheckProbe.checkHealth(port);
        if (health && health.ready) {
          return health;
        }
      } catch (_) {
        // Retry
      }
      await new Promise((resolve) => setTimeout(resolve, delay));
      delay = Math.min(delay * 1.5, 1000);
    }
    throw new Error(`Agent backend health check timed out after ${timeoutMs}ms on port ${port}`);
  }

  public static checkHealth(port: number): Promise<HealthStatus> {
    return new Promise((resolve, reject) => {
      const req = http.get(`http://127.0.0.1:${port}/health`, (res) => {
        if (res.statusCode !== 200) {
          reject(new Error(`HTTP ${res.statusCode}`));
          return;
        }
        let body = '';
        res.on('data', (chunk) => (body += chunk));
        res.on('end', () => {
          try {
            const json = JSON.parse(body);
            resolve(json);
          } catch (err) {
            reject(err);
          }
        });
      });

      req.on('error', (err) => reject(err));
      req.setTimeout(2000, () => {
        req.destroy();
        reject(new Error('Health check request timeout'));
      });
    });
  }
}
