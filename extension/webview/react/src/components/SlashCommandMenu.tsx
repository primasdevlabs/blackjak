import React, { useEffect, useState } from 'react';

export interface CommandItem {
  name: string;
  description: string;
  shortcut?: string;
}

interface SlashCommandMenuProps {
  query: string;
  onSelect: (command: CommandItem) => void;
  onClose: () => void;
}

export const DEFAULT_COMMANDS: CommandItem[] = [
  { name: '/compact', description: 'Compact context' },
  { name: '/context', description: 'Context status' },
  { name: '/clear', description: 'New context' },
  { name: '/summarize', description: 'Summarize task' },
  { name: '/plan', description: 'Create plan' },
  { name: '/model', description: 'Change model' },
  { name: '/effort', description: 'Change effort' },
  { name: '/agents', description: 'Show subagents' },
  { name: '/queue', description: 'Show queue' },
  { name: '/changes', description: 'Show changes' },
  { name: '/files', description: 'Attach files' },
  { name: '/undo', description: 'Undo last step' },
  { name: '/stop', description: 'Stop agent' },
];

export const SlashCommandMenu: React.FC<SlashCommandMenuProps> = ({ query, onSelect, onClose }) => {
  const [selectedIndex, setSelectedIndex] = useState(0);

  const cleanQuery = query.toLowerCase().trim();
  const filtered = DEFAULT_COMMANDS.filter((cmd) =>
    cmd.name.toLowerCase().includes(cleanQuery) ||
    cmd.description.toLowerCase().includes(cleanQuery.replace('/', ''))
  );

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (filtered.length === 0) return;

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev + 1) % filtered.length);
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev - 1 + filtered.length) % filtered.length);
      } else if (e.key === 'Enter' || e.key === 'Tab') {
        e.preventDefault();
        if (filtered[selectedIndex]) {
          onSelect(filtered[selectedIndex]);
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown, true);
    return () => window.removeEventListener('keydown', handleKeyDown, true);
  }, [filtered, selectedIndex, onSelect, onClose]);

  if (filtered.length === 0) return null;

  return (
    <div className="slash-command-menu">
      <div className="slash-header">
        <span>┌───────────────────────────────┐</span>
        <span>│ {query || '/'}</span>
        <span>├───────────────────────────────┤</span>
      </div>
      <div className="slash-list">
        {filtered.map((cmd, idx) => (
          <div
            key={cmd.name}
            className={`slash-item ${idx === selectedIndex ? 'selected' : ''}`}
            onClick={() => onSelect(cmd)}
            onMouseEnter={() => setSelectedIndex(idx)}
          >
            <span className="slash-name">{cmd.name}</span>
            <span className="slash-desc">{cmd.description}</span>
          </div>
        ))}
      </div>
      <div className="slash-footer">
        <span>└───────────────────────────────┘</span>
      </div>
    </div>
  );
};
