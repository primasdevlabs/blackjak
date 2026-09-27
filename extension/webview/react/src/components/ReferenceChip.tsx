import React from 'react';
import { DocumentTextIcon, FolderIcon, XMarkIcon } from '@heroicons/react/24/outline';
import { WorkspaceReference } from '../types/events';

interface ReferenceChipProps {
  reference: WorkspaceReference;
  onRemove: (raw: string) => void;
}

/**
 * Renders a visual chip for a parsed @workspace reference.
 * Shows file/folder icon + path, with a dismiss button.
 */
export const ReferenceChip: React.FC<ReferenceChipProps> = ({ reference, onRemove }) => {
  const label = reference.line ? `${reference.path}:${reference.line}` : reference.path;

  return (
    <span className="reference-chip">
      {reference.type === 'folder' ? (
        <FolderIcon className="reference-chip-icon" />
      ) : (
        <DocumentTextIcon className="reference-chip-icon" />
      )}
      <span className="reference-chip-label">{label}</span>
      <button
        className="reference-chip-remove"
        onClick={(e) => {
          e.preventDefault();
          e.stopPropagation();
          onRemove(reference.raw);
        }}
        title="Remove reference"
      >
        <XMarkIcon className="reference-chip-dismiss-icon" />
      </button>
    </span>
  );
};

interface ReferenceChipListProps {
  references: WorkspaceReference[];
  onRemove: (raw: string) => void;
}

/**
 * Renders a horizontal list of reference chips parsed from composer input.
 * Only displayed when the composer contains @path tokens.
 */
export const ReferenceChipList: React.FC<ReferenceChipListProps> = ({ references, onRemove }) => {
  if (!references || references.length === 0) return null;

  return (
    <div className="reference-chip-list">
      {references.map((ref, idx) => (
        <ReferenceChip key={`${ref.raw}-${idx}`} reference={ref} onRemove={onRemove} />
      ))}
    </div>
  );
};
