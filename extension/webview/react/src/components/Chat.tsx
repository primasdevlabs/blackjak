import React, { useEffect, useMemo, useRef } from 'react';
import { ChatMessage } from '../state/agentStore';
import { Message } from './Message';

interface ChatProps {
  messages: ChatMessage[];
  disabled?: boolean;
  compact?: boolean;
}

function dedupeAdjacent(messages: ChatMessage[]): ChatMessage[] {
  const out: ChatMessage[] = [];
  for (const m of messages) {
    const prev = out[out.length - 1];
    if (
      prev &&
      prev.sender === m.sender &&
      prev.text.trim() === m.text.trim() &&
      (prev.kind || 'text') === (m.kind || 'text')
    ) {
      continue;
    }
    out.push(m);
  }
  return out;
}

export const Chat: React.FC<ChatProps> = ({ messages, compact = false }) => {
  const bottomRef = useRef<HTMLDivElement>(null);

  const visible = useMemo(() => {
    const base = dedupeAdjacent(messages);
    if (!compact) return base;
    return base.filter((m) => {
      if (m.sender === 'user') return true;
      if (m.sender === 'agent' && (m.text?.length || 0) < 400) return true;
      if (/compact|error|failed|complete|paused/i.test(m.text || '')) return true;
      return false;
    });
  }, [messages, compact]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [visible]);

  return (
    <div className={`conversation-scroll-area ${compact ? 'compact-chat' : ''}`}>
      {visible.length === 0 ? (
        <div className="chat-empty">Describe a task to begin.</div>
      ) : (
        visible.map((msg, i) => {
          const prev = visible[i - 1];
          const showSender =
            msg.sender !== 'user' &&
            (!prev || prev.sender === 'user' || prev.kind === 'choice');
          return <Message key={msg.id} message={msg} showSender={showSender} />;
        })
      )}
      <div ref={bottomRef} />
    </div>
  );
};
