import React, { useState, useCallback } from 'react';
import { ArrowUpTrayIcon } from '@heroicons/react/24/outline';

interface DropZoneProps {
  onFileDrop: (path: string, type: 'file' | 'folder') => void;
  children: React.ReactNode;
}

/**
 * Full-container drag-and-drop overlay. When files or folders are dragged
 * over the app, shows a visual overlay indicating the drop zone. On drop,
 * converts the DataTransfer items into structured workspace references.
 */
export const DropZone: React.FC<DropZoneProps> = ({ onFileDrop, children }) => {
  const [isDragging, setIsDragging] = useState(false);
  const [dragCounter, setDragCounter] = useState(0);

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragCounter((prev) => {
      const next = prev + 1;
      if (next === 1) setIsDragging(true);
      return next;
    });
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragCounter((prev) => {
      const next = prev - 1;
      if (next === 0) setIsDragging(false);
      return next;
    });
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragging(false);
      setDragCounter(0);

      // Try webkitGetAsEntry for folder detection
      const items = e.dataTransfer.items;
      if (items && items.length > 0) {
        for (let i = 0; i < items.length; i++) {
          const item = items[i];
          // webkitGetAsEntry is supported in most modern browsers & VS Code webviews
          const entry = (item as any).webkitGetAsEntry?.();
          if (entry) {
            const isFolder = entry.isDirectory;
            const path = (e.dataTransfer.files[i] as any)?.path || entry.fullPath || entry.name;
            onFileDrop(path, isFolder ? 'folder' : 'file');
          } else if (e.dataTransfer.files[i]) {
            const file = e.dataTransfer.files[i];
            const path = (file as any).path || file.name;
            onFileDrop(path, 'file');
          }
        }
        return;
      }

      // Fallback: plain File objects
      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
        for (let i = 0; i < e.dataTransfer.files.length; i++) {
          const file = e.dataTransfer.files[i];
          const path = (file as any).path || file.name;
          onFileDrop(path, 'file');
        }
      }
    },
    [onFileDrop]
  );

  return (
    <div
      className="drop-zone-container"
      onDragEnter={handleDragEnter}
      onDragLeave={handleDragLeave}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
    >
      {children}
      {isDragging && (
        <div className="drop-zone-overlay">
          <div className="drop-zone-content">
            <ArrowUpTrayIcon className="drop-zone-icon" />
            <div className="drop-zone-text">Drop files or folders to attach</div>
            <div className="drop-zone-hint">They will be added as workspace context</div>
          </div>
        </div>
      )}
    </div>
  );
};
