import React, { useEffect, useMemo, useState } from 'react';
import { CodeBracketIcon, DocumentIcon, FolderIcon } from '@heroicons/react/24/outline';
import { apiClient } from '../api/client';

export interface MentionItem {
  type: 'file' | 'folder' | 'symbol';
  path: string;
  name: string;
  kind?: string;
  line?: number;
}

interface MentionPickerProps {
  query: string; // text after @
  onSelect: (item: MentionItem) => void;
  onClose: () => void;
}

export const MentionPicker: React.FC<MentionPickerProps> = ({ query, onSelect, onClose }) => {
  const [items, setItems] = useState<MentionItem[]>([]);
  const [active, setActive] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setLoading(true);
      try {
        const [filesRes, symRes] = await Promise.all([
          fetch(`${apiClient.getBaseUrl()}/api/workspace/files?q=${encodeURIComponent(query)}&limit=24`),
          fetch(`${apiClient.getBaseUrl()}/api/workspace/symbols?q=${encodeURIComponent(query)}&limit=24`),
        ]);
        const filesJson = await filesRes.json();
        const symJson = await symRes.json();
        const files: MentionItem[] = filesJson.success && Array.isArray(filesJson.data) ? filesJson.data : [];
        const symbols: MentionItem[] =
          symJson.success && Array.isArray(symJson.data)
            ? symJson.data.map((s: any) => ({
                type: 'symbol' as const,
                path: s.path,
                name: s.name,
                kind: s.kind,
                line: s.line,
              }))
            : [];
        if (!cancelled) {
          // Prefer symbols when query looks like an identifier.
          const ident = /^[A-Za-z_][\w]*$/.test(query);
          setItems(ident ? [...symbols, ...files] : [...files, ...symbols]);
          setActive(0);
        }
      } catch {
        if (!cancelled) setItems([]);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    const t = setTimeout(load, 120);
    return () => {
      cancelled = true;
      clearTimeout(t);
    };
  }, [query]);

  const filtered = useMemo(() => items, [items]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      } else if (e.key === 'ArrowDown') {
        e.preventDefault();
        setActive((i) => Math.min(i + 1, filtered.length - 1));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setActive((i) => Math.max(i - 1, 0));
      } else if (e.key === 'Enter' && filtered[active]) {
        e.preventDefault();
        onSelect(filtered[active]);
      }
    };
    window.addEventListener('keydown', onKey, true);
    return () => window.removeEventListener('keydown', onKey, true);
  }, [filtered, active, onSelect, onClose]);

  const iconFor = (item: MentionItem) => {
    if (item.type === 'folder') return <FolderIcon className="icon-sm" />;
    if (item.type === 'symbol') return <CodeBracketIcon className="icon-sm" />;
    return <DocumentIcon className="icon-sm" />;
  };

  return (
    <div className="mention-picker" role="listbox">
      <div className="mention-picker-header">@ context · files & symbols</div>
      {loading && <div className="mention-picker-empty">Searching…</div>}
      {!loading && filtered.length === 0 && (
        <div className="mention-picker-empty">No matches</div>
      )}
      {filtered.map((item, i) => (
        <button
          key={`${item.type}:${item.path}:${item.name}`}
          className={`mention-picker-item ${i === active ? 'active' : ''}`}
          onMouseEnter={() => setActive(i)}
          onClick={() => onSelect(item)}
        >
          {iconFor(item)}
          <span className="mention-picker-name">{item.name}</span>
          <span className="mention-picker-path">
            {item.type === 'symbol' && item.kind ? `${item.kind} · ` : ''}
            {item.path}
          </span>
        </button>
      ))}
    </div>
  );
};
