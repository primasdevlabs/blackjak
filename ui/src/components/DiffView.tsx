import React from 'react';

interface DiffViewProps {
  diffText?: string;
  filePath?: string;
  onClose?: () => void;
}

export const DiffView: React.FC<DiffViewProps> = ({ diffText, filePath, onClose }) => {
  if (!diffText) return null;

  const lines = diffText.split('\n');

  return (
    <div className="diff-container">
      <div className="diff-header">
        <span className="diff-title">Diff: {filePath || 'Modified File'}</span>
        {onClose && (
          <button className="diff-close-btn" onClick={onClose}>
            ✕
          </button>
        )}
      </div>
      <div className="diff-content">
        {lines.map((line, idx) => {
          let lineClass = 'diff-normal';
          if (line.startsWith('+')) lineClass = 'diff-add';
          else if (line.startsWith('-')) lineClass = 'diff-del';

          return (
            <div key={idx} className={`diff-line ${lineClass}`}>
              <span className="line-num">{idx + 1}</span>
              <span className="line-text">{line}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
