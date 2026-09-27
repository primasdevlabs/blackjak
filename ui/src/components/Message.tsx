import React from 'react';
import { ChatMessage } from '../state/agentStore';
import { FileReferences } from './FileReferences';

interface MessageProps {
  message: ChatMessage;
}

export const Message: React.FC<MessageProps> = ({ message }) => {
  const isUser = message.sender === 'user';
  const senderLabel = isUser ? 'YOU' : message.sender ? message.sender.toUpperCase() : 'ORCHESTRATOR';

  return (
    <div className="message-editorial">
      <div className="message-sender">{senderLabel}</div>
      <div className="message-body">
        <FileReferences text={message.text} />
      </div>
      <div className="message-divider" />
    </div>
  );
};
