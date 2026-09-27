import React from 'react';
import { Subagent, FileChange } from '../types/events';
import { XMarkIcon, UserIcon, DocumentCheckIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline';

interface AgentResultProps {
  subagent: Subagent | null;
  fileChanges: FileChange[];
  onClose: () => void;
}

export const AgentResult: React.FC<AgentResultProps> = ({ subagent, fileChanges, onClose }) => {
  if (!subagent) return null;

  const agentChanges = fileChanges.filter((fc) => fc.agentId === subagent.id);

  return (
    <div className="agent-result-modal">
      <div className="agent-result-content">
        <div className="result-header">
          <div className="title-area">
            <UserIcon className="icon header-icon" />
            <span className="result-title">Subagent: {subagent.role.toUpperCase()}</span>
            <span className={`status-badge status-${subagent.status}`}>{subagent.status}</span>
          </div>
          <button className="close-btn" onClick={onClose}>
            <XMarkIcon className="icon" />
          </button>
        </div>

        <div className="result-body">
          <div className="info-group">
            <label>Task:</label>
            <p>{subagent.task}</p>
          </div>

          {subagent.result && (
            <div className="info-group">
              <label>Result:</label>
              <div className="result-box">
                <DocumentCheckIcon className="icon icon-success" />
                <span>{subagent.result}</span>
              </div>
            </div>
          )}

          {subagent.error && (
            <div className="info-group">
              <label>Error:</label>
              <div className="error-box">
                <ExclamationCircleIcon className="icon icon-error" />
                <span>{subagent.error}</span>
              </div>
            </div>
          )}

          {subagent.findings && subagent.findings.length > 0 && (
            <div className="info-group">
              <label>Findings ({subagent.findings.length}):</label>
              <div className="findings-container">
                {subagent.findings.map((f, idx) => (
                  <div key={idx} className="finding-card">
                    <span className="file-link">{f.file}{f.line ? `:${f.line}` : ''}</span>
                    <span className="file-desc">{f.message}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {agentChanges.length > 0 && (
            <div className="info-group">
              <label>Files Touched ({agentChanges.length}):</label>
              <div className="touched-files">
                {agentChanges.map((fc, idx) => (
                  <div key={idx} className="touched-file">
                    <span className={`change-badge type-${fc.type}`}>{fc.type}</span>
                    <span className="file-path">{fc.path}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
