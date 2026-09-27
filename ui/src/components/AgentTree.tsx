import React from 'react';
import { Subagent } from '../types/events';

interface AgentTreeProps {
  subagents: Subagent[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}

export const AgentTree: React.FC<AgentTreeProps> = ({ subagents, selectedId, onSelect }) => {
  return (
    <div style={{ margin: '8px 0' }}>
      <div className="section-label">AGENTS</div>
      <div className="font-mono" style={{ padding: '0 16px', fontSize: '11.5px', color: 'var(--text-secondary)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '4px' }}>
          <span className="status-dot working" style={{ width: '6px', height: '6px' }} />
          <span>Orchestrator</span>
        </div>

        {subagents.length === 0 ? (
          <div style={{ paddingLeft: '12px', color: 'var(--text-muted)', fontSize: '11px', marginTop: '2px' }}>
            └ Idle
          </div>
        ) : (
          subagents.map((sub, idx) => {
            const isLast = idx === subagents.length - 1;
            const prefix = isLast ? '└─ ' : '├─ ';
            const isSelected = sub.id === selectedId;

            return (
              <div
                key={sub.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '6px',
                  padding: '2px 0',
                  cursor: 'pointer',
                  color: isSelected ? 'var(--white)' : 'var(--text-secondary)',
                }}
                onClick={() => onSelect(sub.id)}
              >
                <span style={{ color: 'var(--text-disabled)' }}>{prefix}</span>
                <span style={{ fontWeight: isSelected ? 600 : 400, textTransform: 'capitalize' }}>{sub.role}</span>
                {sub.status === 'running' && (
                  <span style={{ fontSize: '10px', color: 'var(--text-muted)', marginLeft: 'auto' }}>●</span>
                )}
                {sub.status === 'completed' && (
                  <span style={{ fontSize: '10px', color: 'var(--text-muted)', marginLeft: 'auto' }}>✓</span>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
