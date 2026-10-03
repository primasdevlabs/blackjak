import React, { useState } from 'react';
import { ChevronDownIcon, ChevronRightIcon } from '@heroicons/react/24/outline';
import { agentHost } from '../host';

interface FilesReviewBarProps {
  filesRead: string[];
  filePaths: string[];
  onReview: () => void;
  reviewDisabled?: boolean;
}

/** Files count + Review — sits just above the composer, not in the activity feed. */
export const FilesReviewBar: React.FC<FilesReviewBarProps> = ({
  filesRead,
  filePaths,
  onReview,
  reviewDisabled,
}) => {
  const [filesOpen, setFilesOpen] = useState(false);
  const uniqueFiles = [...new Set([...filePaths, ...filesRead])];

  return (
    <div className="composer-files-bar">
      <button type="button" className="activity-files-toggle" onClick={() => setFilesOpen((v) => !v)}>
        {filesOpen ? <ChevronDownIcon className="icon-xs" /> : <ChevronRightIcon className="icon-xs" />}
        <span>{uniqueFiles.length} Files</span>
      </button>
      <button
        type="button"
        className="activity-review-btn"
        onClick={onReview}
        disabled={reviewDisabled ?? filePaths.length === 0}
      >
        Review
      </button>
      {filesOpen && uniqueFiles.length > 0 && (
        <div className="composer-files-list">
          {uniqueFiles.map((p) => (
            <button key={p} type="button" className="activity-files-item" onClick={() => agentHost.openFile(p)}>
              {p}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
