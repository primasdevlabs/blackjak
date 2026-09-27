import React from 'react';
import { CommandLineIcon } from '@heroicons/react/24/outline';

interface TerminalProps {
  logs: string[];
}

export const Terminal: React.FC<TerminalProps> = ({ logs }) => {
  if (logs.length === 0) return null;

  return (
    <div className="card terminal-card">
      <div className="card-header">
        <CommandLineIcon className="icon header-icon" />
        <span className="card-title">Terminal</span>
      </div>
      <div className="terminal-logs">
        {logs.map((log, idx) => (
          <pre key={idx} className="terminal-line">
            {log}
          </pre>
        ))}
      </div>
    </div>
  );
};
