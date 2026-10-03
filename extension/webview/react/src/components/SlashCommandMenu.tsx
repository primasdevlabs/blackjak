import React, { useEffect, useState } from 'react';
import { apiClient } from '../api/client';

export interface CommandItem {
  name: string;
  description: string;
  shortcut?: string;
  group?: 'command' | 'skill';
}

interface SlashCommandMenuProps {
  query: string;
  onSelect: (command: CommandItem) => void;
  onClose: () => void;
}

export const DEFAULT_COMMANDS: CommandItem[] = [
  { name: '/compact', description: 'Compact context', group: 'command' },
  { name: '/context', description: 'Context status', group: 'command' },
  { name: '/clear', description: 'New context', group: 'command' },
  { name: '/summarize', description: 'Summarize task', group: 'command' },
  { name: '/plan', description: 'Plan mode', group: 'command' },
  { name: '/ask', description: 'Ask mode', group: 'command' },
  { name: '/agent', description: 'Agent mode', group: 'command' },
  { name: '/model', description: 'Change model', group: 'command' },
  { name: '/effort', description: 'Change effort', group: 'command' },
  { name: '/agents', description: 'Show subagents', group: 'command' },
  { name: '/activity', description: 'Activity log', group: 'command' },
  { name: '/queue', description: 'Show queue', group: 'command' },
  { name: '/changes', description: 'Show changes', group: 'command' },
  { name: '/files', description: 'Attach files', group: 'command' },
  { name: '/pr', description: 'Draft PR summary from git + agent changes', group: 'command' },
  { name: '/undo', description: 'Undo last step', group: 'command' },
  { name: '/stop', description: 'Stop agent', group: 'command' },
];

export const SlashCommandMenu: React.FC<SlashCommandMenuProps> = ({ query, onSelect, onClose }) => {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [skills, setSkills] = useState<CommandItem[]>([]);

  useEffect(() => {
    fetch(`${apiClient.getBaseUrl()}/api/skills`)
      .then((r) => r.json())
      .then((j) => {
        if (!j.success || !Array.isArray(j.data)) return;
        setSkills(
          j.data
            .filter((s: any) => s.enabled !== false)
            .map((s: any) => ({
              name: `/${s.name}`,
              description: s.description || 'Skill',
              group: 'skill' as const,
            }))
        );
      })
      .catch(() => undefined);
  }, []);

  const cleanQuery = query.toLowerCase().trim();
  const filtered = [...DEFAULT_COMMANDS, ...skills].filter(
    (cmd) =>
      cmd.name.toLowerCase().includes(cleanQuery) ||
      cmd.description.toLowerCase().includes(cleanQuery.replace('/', ''))
  );

  useEffect(() => {
    setSelectedIndex(0);
  }, [query, skills.length]);

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
        <span>Commands & skills</span>
      </div>
      <div className="slash-list">
        {filtered.map((cmd, idx) => (
          <div
            key={cmd.name}
            className={`slash-item ${idx === selectedIndex ? 'selected' : ''}`}
            onMouseEnter={() => setSelectedIndex(idx)}
            onClick={() => onSelect(cmd)}
          >
            <span className="slash-name">{cmd.name}</span>
            <span className="slash-desc">
              {cmd.group === 'skill' ? 'Skill · ' : ''}
              {cmd.description}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};
