import React, { useState } from 'react';
import { Attachment } from '../types/events';
import { PaperClipIcon, DocumentIcon, FolderIcon, XMarkIcon, PlusIcon } from '@heroicons/react/24/outline';

interface AttachmentsProps {
  attachments: Attachment[];
  onAdd: (path: string, type?: 'file' | 'folder') => void;
  onRemove: (id: string) => void;
}

export const Attachments: React.FC<AttachmentsProps> = ({ attachments, onAdd, onRemove }) => {
  const [inputPath, setInputPath] = useState('');
  const [showInput, setShowInput] = useState(false);

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputPath.trim()) return;
    const isFolder = !inputPath.includes('.') || inputPath.endsWith('/');
    onAdd(inputPath.trim(), isFolder ? 'folder' : 'file');
    setInputPath('');
    setShowInput(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      for (let i = 0; i < e.dataTransfer.files.length; i++) {
        const file = e.dataTransfer.files[i];
        onAdd((file as any).path || file.name, 'file');
      }
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  return (
    <div className="attachments-wrapper" onDrop={handleDrop} onDragOver={handleDragOver}>
      <div className="attachments-header">
        <PaperClipIcon className="icon header-icon" />
        <span className="attach-title">Attached Context ({attachments.length})</span>
        <button className="add-attach-btn" onClick={() => setShowInput(!showInput)}>
          <PlusIcon className="icon btn-icon-sm" />
          <span>Attach</span>
        </button>
      </div>

      {showInput && (
        <form className="attach-input-form" onSubmit={handleAdd}>
          <input
            type="text"
            className="attach-path-input"
            placeholder="Path to attach (e.g. internal/auth or auth.go)"
            value={inputPath}
            onChange={(e) => setInputPath(e.target.value)}
          />
          <button type="submit" className="btn btn-approve btn-sm">Add</button>
        </form>
      )}

      {attachments.length > 0 && (
        <div className="attachment-chips">
          {attachments.map((att) => (
            <div key={att.id} className="attach-chip">
              {att.type === 'folder' ? (
                <FolderIcon className="icon chip-icon" />
              ) : (
                <DocumentIcon className="icon chip-icon" />
              )}
              <span className="chip-name">{att.name}</span>
              <button className="remove-chip-btn" onClick={() => onRemove(att.id)}>
                <XMarkIcon className="icon btn-icon-sm" />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
