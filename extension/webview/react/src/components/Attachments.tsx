import React, { useState } from 'react';
import { Attachment } from '../types/events';
import { PaperClipIcon, DocumentIcon, FolderIcon, XMarkIcon, PlusIcon } from '@heroicons/react/24/outline';
import { ImageLightbox } from './ImageLightbox';
import { isImageName, workspacePreviewUrl } from '../utils/attachmentPreview';
import { apiClient } from '../api/client';

interface AttachmentsProps {
  attachments: Attachment[];
  onAdd: (path: string, type?: 'file' | 'folder' | 'image') => void;
  onRemove: (id: string) => void;
}

export const Attachments: React.FC<AttachmentsProps> = ({ attachments, onAdd, onRemove }) => {
  const [inputPath, setInputPath] = useState('');
  const [showInput, setShowInput] = useState(false);
  const [lightbox, setLightbox] = useState<string | null>(null);

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputPath.trim()) return;
    const isFolder = !inputPath.includes('.') || inputPath.endsWith('/');
    onAdd(inputPath.trim(), isFolder ? 'folder' : isImageName(inputPath) ? 'image' : 'file');
    setInputPath('');
    setShowInput(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      for (let i = 0; i < e.dataTransfer.files.length; i++) {
        const file = e.dataTransfer.files[i];
        onAdd((file as any).path || file.name, file.type.startsWith('image/') ? 'image' : 'file');
      }
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  if (attachments.length === 0 && !showInput) {
    return null;
  }

  return (
    <div className="attachments-wrapper" onDrop={handleDrop} onDragOver={handleDragOver}>
      <div className="attachments-header">
        <PaperClipIcon className="icon header-icon" />
        <span className="attach-title">Attachments ({attachments.length})</span>
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
          {attachments.map((att) => {
            const preview =
              att.previewUrl ||
              (att.type === 'image' || isImageName(att.name)
                ? workspacePreviewUrl(apiClient.getBaseUrl(), att.path)
                : undefined);
            return (
              <div key={att.id} className={`attach-chip ${preview ? 'attach-chip-image' : ''}`}>
                {preview ? (
                  <button type="button" className="attach-thumb-btn" onClick={() => setLightbox(preview)}>
                    <img src={preview} alt={att.name} className="attach-thumb" />
                  </button>
                ) : att.type === 'folder' ? (
                  <FolderIcon className="icon chip-icon" />
                ) : (
                  <DocumentIcon className="icon chip-icon" />
                )}
                <span className="chip-name">{att.name}</span>
                <button className="remove-chip-btn" onClick={() => onRemove(att.id)}>
                  <XMarkIcon className="icon btn-icon-sm" />
                </button>
              </div>
            );
          })}
        </div>
      )}
      {lightbox && <ImageLightbox src={lightbox} onClose={() => setLightbox(null)} />}
    </div>
  );
};
