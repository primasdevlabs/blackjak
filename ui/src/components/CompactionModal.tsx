import React from 'react';
import { XMarkIcon } from '@heroicons/react/24/outline';

export interface CompactionData {
  tokensBefore: number;
  tokensAfter: number;
  tokensSaved: number;
  preserved: string[];
  removed: string[];
}

interface CompactionModalProps {
  data: CompactionData;
  onClose: () => void;
}

export const CompactionModal: React.FC<CompactionModalProps> = ({ data, onClose }) => {
  const formatTokens = (n: number) => {
    if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
    return n.toString();
  };

  return (
    <div className="popover-overlay" onClick={onClose}>
      <div className="popover-card" style={{ width: '320px' }} onClick={(e) => e.stopPropagation()}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px', borderBottom: '1px solid var(--border)', paddingBottom: '8px' }}>
          <span style={{ fontWeight: 600, fontSize: '12px', letterSpacing: '0.04em' }}>CONTEXT COMPACTION</span>
          <button className="icon-btn" onClick={onClose}><XMarkIcon className="icon-sm" /></button>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '14px', fontSize: '12px' }}>
          <div>
            <div style={{ color: 'var(--text-muted)', fontSize: '11px', fontWeight: 600, marginBottom: '6px' }}>PRESERVED</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
              {(data.preserved || [
                'Task objective',
                'Implementation decisions',
                'Modified files state',
                'Test results',
                'Subagent findings'
              ]).map((item, idx) => (
                <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: '6px', color: 'var(--text-primary)' }}>
                  <span style={{ color: '#ffffff' }}>✓</span>
                  <span>{item}</span>
                </div>
              ))}
            </div>
          </div>

          <div>
            <div style={{ color: 'var(--text-muted)', fontSize: '11px', fontWeight: 600, marginBottom: '6px' }}>REMOVED</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
              {(data.removed || [
                '12 tool outputs',
                '8 duplicate search results',
                '4 superseded plans'
              ]).map((item, idx) => (
                <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: '6px', color: 'var(--text-secondary)' }}>
                  <span style={{ color: 'var(--text-muted)' }}>–</span>
                  <span>{item}</span>
                </div>
              ))}
            </div>
          </div>

          <div style={{ borderTop: '1px solid var(--border)', paddingTop: '10px', display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '8px', textAlign: 'center' }}>
            <div>
              <div style={{ fontSize: '10px', color: 'var(--text-muted)' }}>Before</div>
              <div style={{ fontWeight: 600, color: 'var(--text-secondary)' }}>{formatTokens(data.tokensBefore || 42100)}</div>
            </div>
            <div>
              <div style={{ fontSize: '10px', color: 'var(--text-muted)' }}>After</div>
              <div style={{ fontWeight: 600, color: '#ffffff' }}>{formatTokens(data.tokensAfter || 11700)}</div>
            </div>
            <div>
              <div style={{ fontSize: '10px', color: 'var(--text-muted)' }}>Saved</div>
              <div style={{ fontWeight: 600, color: '#ffffff' }}>{formatTokens(data.tokensSaved || 30400)}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
