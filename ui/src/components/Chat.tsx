import React, { useState } from 'react';
import { ChatMessage } from '../state/agentStore';
import { Message } from './Message';
import { PaperAirplaneIcon } from '@heroicons/react/24/outline';

interface ChatProps {
  messages: ChatMessage[];
  onSubmit: (prompt: string) => void;
  disabled?: boolean;
}

export const Chat: React.FC<ChatProps> = ({ messages, onSubmit, disabled }) => {
  const [input, setInput] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || disabled) return;
    onSubmit(input.trim());
    setInput('');
  };

  return (
    <div className="chat-container">
      <div className="messages-list">
        {messages.map((msg) => (
          <Message key={msg.id} message={msg} />
        ))}
      </div>

      <form className="chat-input-form" onSubmit={handleSubmit}>
        <input
          type="text"
          className="chat-input"
          placeholder="Ask the agent... (e.g. Fix the failing tests)"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          disabled={disabled}
        />
        <button type="submit" className="chat-submit-btn" disabled={disabled || !input.trim()}>
          <PaperAirplaneIcon className="icon btn-icon" />
        </button>
      </form>
    </div>
  );
};
