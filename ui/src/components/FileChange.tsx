import React from 'react';
import { FileChange as FileChangeType } from '../types/events';

interface FileChangeProps {
  changes: FileChangeType[];
}

export const FileChange: React.FC<FileChangeProps> = ({ changes }) => {
  if (changes.length === 0) return null;

  const handleOpenFile = (path: string) => {
    if (typeof (window as any).vscode !== 'undefined') {
      (window as any).vscode.postMessage({
        command: 'openFile',
        payload: { path },
      });
    }
  };

  const handleShowDiff = (change: FileChangeType) => {
    if (typeof (window as any).vscode !== 'undefined') {
      (window as any).vscode.postMessage({
        command: 'openDiff',
        payload: { path: change.path },
      });
    }
  };

  const getTypeLetter = (type: string) => {
    switch (type.toLowerCase()) {
      case 'modified': case 'm': return 'M';
      case 'added': case 'a': return 'A';
      case 'deleted': case 'd': return 'D';
      case 'renamed': case 'r': return 'R';
      default: return 'M';
    }
  };

  return (
    <div style={{ margin: '12px 0' }}>
      <div className="section-label">CHANGES</div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', padding: '0 4px' }}>
        {changes.map((c, idx) => {
          const letter = getTypeLetter(c.type);
          return (
            <div key={idx} className="file-change-row">
              <span className={`change-badge ${letter}`} onClick={() => handleShowDiff(c)}>
                {letter}
              </span>
              <span className="font-mono" style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} onClick={() => handleOpenFile(c.path)}>
                {c.path}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
