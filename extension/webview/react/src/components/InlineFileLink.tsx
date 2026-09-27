import React from 'react';
import { DocumentTextIcon } from '@heroicons/react/24/outline';
import { agentHost } from '../host';

interface InlineFileLinkProps {
  path: string;
  line?: number;
  displayText?: string;
}

/**
 * Renders a clickable file reference that opens the file in the IDE editor.
 * Supports optional line number navigation. Used inline within agent messages
 * to make @path/to/file:line references interactive.
 */
export const InlineFileLink: React.FC<InlineFileLinkProps> = ({ path, line, displayText }) => {
  const label = displayText || (line ? `${path}:${line}` : path);

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    agentHost.openFile(path, line ? { line } : undefined);
  };

  return (
    <span className="inline-file-link" onClick={handleClick} title={`Open ${path}${line ? ` at line ${line}` : ''}`}>
      <DocumentTextIcon className="inline-file-icon" />
      <span className="inline-file-label">{label}</span>
    </span>
  );
};
