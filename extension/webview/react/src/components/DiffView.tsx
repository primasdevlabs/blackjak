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
  let lineNoOld = 0;
  let lineNoNew = 0;

  return (
    <div className="diff-view">
      {(filePath || onClose) && (
        <div className="diff-view-header">
          <span>{filePath || 'Diff'}</span>
          {onClose && (
            <button type="button" className="diff-view-close" onClick={onClose}>
              <XMarkIcon className="icon-sm" />
            </button>
          )}
        </div>
      )}
      <div className="diff-view-body">
        {lines.map((line, idx) => {
          let kind: 'add' | 'del' | 'meta' | 'ctx' = 'ctx';
          if (line.startsWith('+++') || line.startsWith('---') || line.startsWith('@@')) {
            kind = 'meta';
          } else if (line.startsWith('+')) {
            kind = 'add';
            lineNoNew++;
          } else if (line.startsWith('-')) {
            kind = 'del';
            lineNoOld++;
          } else {
            lineNoOld++;
            lineNoNew++;
          }
          const display = kind === 'add' || kind === 'del' ? line.slice(1) : line;
          const ln =
            kind === 'add' ? lineNoNew : kind === 'del' ? lineNoOld : lineNoNew;
          return (
            <div key={idx} className={`diff-line diff-${kind}`}>
              <span className="diff-ln">{kind === 'meta' ? '' : ln}</span>
              <span className="diff-prefix">{kind === 'add' ? '+' : kind === 'del' ? '-' : ' '}</span>
              <span className="diff-code">{display}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
