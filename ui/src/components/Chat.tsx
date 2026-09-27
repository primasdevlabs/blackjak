import React, { useEffect, useRef } from 'react';
import { ChatMessage } from '../state/agentStore';
import { Message } from './Message';

interface ChatProps {
  messages: ChatMessage[];
  disabled?: boolean;
}

export const Chat: React.FC<ChatProps> = ({ messages }) => {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className="conversation-scroll-area">
      {messages.length === 0 ? (
        <div style={{ color: 'var(--text-muted)', fontSize: '13px', paddingTop: '16px' }}>
          Describe what you'd like the agent to work on below.
        </div>
      ) : (
        messages.map((msg) => <Message key={msg.id} message={msg} />)
      )}
      <div ref={bottomRef} />
    </div>
  );
};
