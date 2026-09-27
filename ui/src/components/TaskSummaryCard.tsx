import React from 'react';

export interface TaskSummaryData {
  objective: string;
  workspace: string;
  completed: string[];
  remaining: string[];
  files: { path: string; status: string }[];
}

interface TaskSummaryCardProps {
  summary: TaskSummaryData;
}

export const TaskSummaryCard: React.FC<TaskSummaryCardProps> = ({ summary }) => {
  return (
    <div style={{ backgroundColor: 'var(--bg-surface)', border: '1px solid var(--border)', borderRadius: '6px', padding: '12px 14px', margin: '12px 0' }}>
      <div style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text-muted)', letterSpacing: '0.06em', marginBottom: '8px' }}>TASK SUMMARY</div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', fontSize: '12px' }}>
        <div>
          <div style={{ fontWeight: 600, color: 'var(--text-primary)', marginBottom: '2px' }}>Objective</div>
          <div style={{ color: 'var(--text-secondary)' }}>{summary.objective || 'Fix token expiration handling'}</div>
        </div>

        {summary.workspace && (
          <div>
            <div style={{ fontWeight: 600, color: 'var(--text-primary)', marginBottom: '2px' }}>Workspace</div>
            <div style={{ color: 'var(--text-secondary)' }}>{summary.workspace}</div>
          </div>
        )}

        <div>
          <div style={{ fontWeight: 600, color: 'var(--text-primary)', marginBottom: '4px' }}>Completed</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '3px' }}>
            {(summary.completed || [
              'Located token validation',
              'Updated expiration comparison',
              'Added middleware handling'
            ]).map((item, idx) => (
              <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: '6px', color: 'var(--text-primary)' }}>
                <span>✓</span>
                <span>{item}</span>
              </div>
            ))}
          </div>
        </div>

        <div>
          <div style={{ fontWeight: 600, color: 'var(--text-primary)', marginBottom: '4px' }}>Remaining</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '3px' }}>
            {(summary.remaining || [
              'Add regression tests',
              'Run auth test suite'
            ]).map((item, idx) => (
              <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: '6px', color: 'var(--text-secondary)' }}>
                <span>□</span>
                <span>{item}</span>
              </div>
            ))}
          </div>
        </div>

        {summary.files && summary.files.length > 0 && (
          <div>
            <div style={{ fontWeight: 600, color: 'var(--text-primary)', marginBottom: '4px' }}>Files</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '3px' }}>
              {summary.files.map((f, idx) => (
                <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <span className={`change-badge ${f.status}`}>{f.status}</span>
                  <span style={{ color: 'var(--text-primary)' }}>{f.path}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
