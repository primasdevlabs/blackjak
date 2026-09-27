import React from 'react';
import { XMarkIcon } from '@heroicons/react/24/outline';

export interface ContextComposition {
  system: number;
  task: number;
  conversation: number;
  files: number;
  toolResults: number;
  agentSummaries: number;
  total: number;
  limit: number;
  pressure: number;
}

interface ContextStatusModalProps {
  composition: ContextComposition;
  onClose: () => void;
  onCompactNow: () => void;
}

export const ContextStatusModal: React.FC<ContextStatusModalProps> = ({ composition, onClose, onCompactNow }) => {
  const formatK = (tokens: number) => {
    return (tokens / 1000).toFixed(1) + 'k';
  };

  const rows = [
    { label: 'System', tokens: composition.system },
    { label: 'Task', tokens: composition.task },
    { label: 'Conversation', tokens: composition.conversation },
    { label: 'Files', tokens: composition.files },
    { label: 'Tool results', tokens: composition.toolResults },
    { label: 'Agent summaries', tokens: composition.agentSummaries },
  ];

  return (
    <div className="popover-overlay" onClick={onClose}>
      <div className="popover-card" style={{ width: '300px' }} onClick={(e) => e.stopPropagation()}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px', borderBottom: '1px solid var(--border)', paddingBottom: '8px' }}>
          <span style={{ fontWeight: 600, fontSize: '12px', letterSpacing: '0.04em' }}>CONTEXT Composition</span>
          <button className="icon-btn" onClick={onClose}><XMarkIcon className="icon-sm" /></button>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', fontSize: '12px' }}>
          {rows.map((r) => (
            <div key={r.label} style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-secondary)' }}>
              <span>{r.label}</span>
              <span style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{formatK(r.tokens)}</span>
            </div>
          ))}

          <div style={{ borderTop: '1px solid var(--border)', paddingTop: '8px', marginTop: '4px', display: 'flex', justifyContent: 'space-between', fontWeight: 600 }}>
            <span>Total</span>
            <span style={{ color: '#ffffff' }}>{formatK(composition.total)}</span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-muted)' }}>
            <span>Limit</span>
            <span>{formatK(composition.limit)}</span>
          </div>

          <div style={{ margin: '10px 0 4px 0' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: 'var(--text-muted)', marginBottom: '4px' }}>
              <span>Pressure</span>
              <span style={{ fontWeight: 600, color: composition.pressure >= 80 ? '#f87171' : 'var(--text-primary)' }}>{composition.pressure}%</span>
            </div>
            <div style={{ width: '100%', height: '4px', backgroundColor: 'var(--bg-elevated)', borderRadius: '2px', overflow: 'hidden' }}>
              <div
                style={{
                  width: `${Math.min(composition.pressure, 100)}%`,
                  height: '100%',
                  backgroundColor: composition.pressure >= 80 ? '#ef4444' : '#ffffff',
                  transition: 'width 0.3s ease',
                }}
              />
            </div>
          </div>

          <button
            className="send-btn"
            style={{ marginTop: '8px', width: '100%', justifyContent: 'center' }}
            onClick={() => {
              onCompactNow();
              onClose();
            }}
          >
            Compact Now (/compact)
          </button>
        </div>
      </div>
    </div>
  );
};
