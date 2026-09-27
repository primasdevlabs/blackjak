import React from 'react';
import { ChatMessage } from '../state/agentStore';
import { FileReferences } from './FileReferences';

interface MessageProps {
  message: ChatMessage;
}

export const Message: React.FC<MessageProps> = ({ message }) => {
  const isUser = message.sender === 'user';

  return (
    <div className={`message-bubble message-${message.sender}`}>
      <div className="message-sender">{isUser ? 'User' : 'Agent'}</div>
      <div className="message-text">
        <FileReferences text={message.text} />
      </div>
    </div>
  );
};
