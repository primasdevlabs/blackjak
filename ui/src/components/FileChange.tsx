import React from 'react';
import { FileChange as FileChangeType } from '../types/events';
import { FolderOpenIcon, DocumentMagnifyingGlassIcon } from '@heroicons/react/24/outline';

interface FileChangeProps {
  changes: FileChangeType[];
}

export const FileChange: React.FC<FileChangeProps> = ({ changes }) => {
  if (changes.length === 0) return null;

  const handleOpenFile = (path: string) => {
    if (typeof (window as any).vscode !== 'undefined') {
      (window as any).vscode.postMessage({
        command: 'openFile',
        path,
      });
    } else {
      console.log('Open file:', path);
    }
  };

  const handleShowDiff = (change: FileChangeType) => {
    if (typeof (window as any).vscode !== 'undefined') {
      (window as any).vscode.postMessage({
        command: 'showDiff',
        path: change.path,
        diff: change.diff,
      });
    } else {
      console.log('Show diff:', change.path, change.diff);
    }
  };

  return (
    <div className="card file-card">
      <div className="card-header">
        <FolderOpenIcon className="icon header-icon" />
        <span className="card-title">File Changes</span>
      </div>
      <div className="file-list">
        {changes.map((c, idx) => (
          <div key={idx} className={`file-item file-${c.type}`}>
            <span className="file-type-badge">{c.type.toUpperCase()}</span>
            <span className="file-path" onClick={() => handleOpenFile(c.path)}>
              {c.path}
            </span>
            {c.diff && (
              <button className="diff-btn" onClick={() => handleShowDiff(c)}>
                <DocumentMagnifyingGlassIcon className="icon btn-icon-sm" />
                <span>View Diff</span>
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};
