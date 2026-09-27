import React from 'react';
import { FileChange as FileChangeType } from '../types/events';
import { agentHost } from '../host';

interface FileChangeProps {
  changes: (FileChangeType & { isPreExisting?: boolean; role?: string })[];
}

export const FileChange: React.FC<FileChangeProps> = ({ changes }) => {
  if (!changes || changes.length === 0) return null;

  const handleOpenFile = (path: string) => {
    agentHost.openFile(path);
  };

  const handleShowDiff = (change: FileChangeType) => {
    agentHost.openDiff(change.path);
  };

  const handleOpenAllChanged = () => {
    changes.forEach((c) => agentHost.openFile(c.path));
  };

  const getTypeLetter = (type: string) => {
    switch (type.toLowerCase()) {
      case 'modified': case 'm': return 'M';
      case 'added': case 'created': case 'a': return 'A';
      case 'deleted': case 'd': return 'D';
      case 'renamed': case 'r': return 'R';
      default: return 'M';
    }
  };

  const taskChanges = changes.filter((c) => !c.isPreExisting);
  const preExistingChanges = changes.filter((c) => c.isPreExisting);

  return (
    <div style={{ margin: '12px 0' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
        <div className="section-label">CHANGES · {changes.length}</div>
        <button className="chip-btn" style={{ fontSize: '10px', padding: '2px 6px' }} onClick={handleOpenAllChanged}>
          Open all
        </button>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', padding: '0 4px' }}>
        {taskChanges.length > 0 && (
          <div>
            <div style={{ fontSize: '10px', fontWeight: 600, color: 'var(--text-muted)', letterSpacing: '0.05em', marginBottom: '4px' }}>THIS TASK</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
              {taskChanges.map((c, idx) => {
                const letter = getTypeLetter(c.type);
                return (
                  <div key={idx} className="file-change-row" style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
                    <span className={`change-badge ${letter}`} onClick={() => handleShowDiff(c)}>
                      {letter}
                    </span>
                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: '12px' }} onClick={() => handleOpenFile(c.path)}>
                      {c.path}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {preExistingChanges.length > 0 && (
          <div>
            <div style={{ fontSize: '10px', fontWeight: 600, color: 'var(--text-muted)', letterSpacing: '0.05em', marginBottom: '4px', marginTop: '6px' }}>PRE-EXISTING</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
              {preExistingChanges.map((c, idx) => {
                const letter = getTypeLetter(c.type);
                return (
                  <div key={idx} className="file-change-row" style={{ display: 'flex', alignItems: 'center', gap: '8px', opacity: 0.7, cursor: 'pointer' }}>
                    <span className={`change-badge ${letter}`} onClick={() => handleShowDiff(c)}>
                      {letter}
                    </span>
                    <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontSize: '12px' }} onClick={() => handleOpenFile(c.path)}>
                      {c.path}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
