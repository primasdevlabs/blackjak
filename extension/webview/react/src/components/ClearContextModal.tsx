import React from 'react';
import { XMarkIcon } from '@heroicons/react/24/outline';

interface ClearContextModalProps {
  onConfirm: () => void;
  onClose: () => void;
}

export const ClearContextModal: React.FC<ClearContextModalProps> = ({ onConfirm, onClose }) => {
  return (
    <div className="popover-overlay" onClick={onClose}>
      <div className="popover-card" style={{ width: '300px' }} onClick={(e) => e.stopPropagation()}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px', borderBottom: '1px solid var(--border)', paddingBottom: '8px' }}>
          <span style={{ fontWeight: 600, fontSize: '12px', letterSpacing: '0.04em' }}>CLEAR CONTEXT</span>
          <button className="icon-btn" onClick={onClose}><XMarkIcon className="icon-sm" /></button>
        </div>

        <div style={{ fontSize: '12px', color: 'var(--text-secondary)', lineHeight: '1.5', marginBottom: '16px' }}>
          <p style={{ fontWeight: 600, color: 'var(--text-primary)', marginBottom: '6px' }}>Clear context?</p>
          <p>This will start a fresh conversation context. The workspace and files will remain unchanged.</p>
        </div>

        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
          <button className="chip-btn" onClick={onClose}>Cancel</button>
          <button
            className="send-btn"
            style={{ backgroundColor: '#ffffff', color: '#000000' }}
            onClick={() => {
              onConfirm();
              onClose();
            }}
          >
            Start fresh
          </button>
        </div>
      </div>
    </div>
  );
};
