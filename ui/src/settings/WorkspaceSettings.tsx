import React from 'react';
import { FolderIcon, ShieldCheckIcon } from '@heroicons/react/24/outline';

interface WorkspaceSettingsProps {
  workspacePath: string;
}

export const WorkspaceSettings: React.FC<WorkspaceSettingsProps> = ({ workspacePath }) => {
  return (
    <div className="card workspace-settings-card">
      <div className="card-header">
        <FolderIcon className="icon header-icon" />
        <span className="card-title">Workspace Boundary & Security</span>
      </div>

      <div className="form-group-list">
        <div className="form-row">
          <label>Active Workspace Root:</label>
          <div className="path-box">
            <FolderIcon className="icon" />
            <span>{workspacePath || 'No active workspace'}</span>
          </div>
        </div>

        <div className="security-notice">
          <ShieldCheckIcon className="icon icon-success" />
          <span>
            Workspace sandbox isolation is active. All file operations, command executions, and subagent scopes are strictly bounded to this workspace path.
          </span>
        </div>
      </div>
    </div>
  );
};
