import React from 'react';
import { ChatMessage } from '../state/agentStore';
import { FileReferences } from './FileReferences';

interface MessageProps {
  message: ChatMessage;
}

export const Message: React.FC<MessageProps> = ({ message }) => {
  const isUser = message.sender === 'user';
  const senderLabel =
    !isUser && message.sender && message.sender !== 'agent'
      ? message.sender.charAt(0).toUpperCase() + message.sender.slice(1)
      : 'BlackJak';

  return (
    <div className={`message-editorial ${isUser ? 'message-user' : ''}`}>
      {!isUser && <div className="message-sender">{senderLabel}</div>}
      <div className="message-body">
        <FileReferences text={message.text} />
      </div>
      <div className="message-divider" />
    </div>
  );
};
