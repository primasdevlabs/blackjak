import React from 'react';
import { DocumentTextIcon } from '@heroicons/react/24/outline';
import { agentHost } from '../host';

interface FileReferencesProps {
  text: string;
}

export const FileReferences: React.FC<FileReferencesProps> = ({ text }) => {
  // Regex matching @path:line or standard relative file paths like internal/auth/middleware.go:84
  const refPattern = /@?(?:(file|folder):)?([a-zA-Z0-9_\-./]+\.[a-zA-Z0-9]+(?::\d+)?)/g;

  const parts: React.ReactNode[] = [];
  let lastIndex = 0;
  let match: RegExpExecArray | null;

  const handleOpenReference = (fullRef: string) => {
    let cleanPath = fullRef.replace(/^@/, '').replace(/^(file|folder):/, '');
    let line = 0;

    if (cleanPath.includes(':')) {
      const idx = cleanPath.lastIndexOf(':');
      line = parseInt(cleanPath.substring(idx + 1), 10) || 0;
      cleanPath = cleanPath.substring(0, idx);
    }

    agentHost.openFile(cleanPath, { line });
  };

  while ((match = refPattern.exec(text)) !== null) {
    const matchStart = match.index;
    const matchText = match[0];
    const targetPath = match[2];

    if (matchStart > lastIndex) {
      parts.push(text.substring(lastIndex, matchStart));
    }

    parts.push(
      <span
        key={matchStart}
        className="interactive-file-ref"
        onClick={() => handleOpenReference(targetPath)}
        title={`Click to open ${targetPath} in editor`}
      >
        <DocumentTextIcon className="icon inline-ref-icon" />
        <span>{matchText}</span>
      </span>
    );

    lastIndex = matchStart + matchText.length;
  }

  if (lastIndex < text.length) {
    parts.push(text.substring(lastIndex));
  }

  return <span className="file-ref-text">{parts}</span>;
};
