import React from 'react';
import { XMarkIcon } from '@heroicons/react/24/outline';

interface DiffViewProps {
  diffText?: string;
  filePath?: string;
  onClose?: () => void;
}

export const DiffView: React.FC<DiffViewProps> = ({ diffText, filePath, onClose }) => {
  if (!diffText) return null;

  const lines = diffText.split('\n');

  return (
    <div style={{ backgroundColor: 'var(--bg-surface)', border: '1px solid var(--border)', borderRadius: '4px', overflow: 'hidden', margin: '8px 0' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '6px 12px', borderBottom: '1px solid var(--border)', backgroundColor: 'var(--bg-elevated)' }}>
        <span style={{ fontSize: '11px', color: 'var(--text-secondary)' }}>{filePath || 'Diff View'}</span>
        {onClose && (
          <button style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }} onClick={onClose}>
            <XMarkIcon className="icon-sm" />
          </button>
        )}
      </div>

      <div style={{ fontSize: '11.5px', lineHeight: '1.5', overflowX: 'auto', padding: '6px 0' }}>
        {lines.map((line, idx) => {
          let bg = 'transparent';
          let color = 'var(--text-primary)';
          let prefix = ' ';

          if (line.startsWith('+')) {
            bg = 'rgba(255, 255, 255, 0.06)';
            color = 'var(--text-primary)';
            prefix = '+';
          } else if (line.startsWith('-')) {
            bg = 'rgba(255, 255, 255, 0.04)';
            color = 'var(--text-muted)';
            prefix = '-';
          }

          return (
            <div key={idx} style={{ backgroundColor: bg, color, padding: '1px 12px', display: 'flex', gap: '8px' }}>
              <span style={{ color: 'var(--text-disabled)', width: '12px', userSelect: 'none' }}>{prefix}</span>
              <span>{line.substring(line.startsWith('+') || line.startsWith('-') ? 1 : 0)}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
