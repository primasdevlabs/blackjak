import React from 'react';
import { ChatMessage } from '../state/agentStore';
import { FileReferences } from './FileReferences';
import { WalkthroughAnswerChip } from './WalkthroughQuestion';

interface MessageProps {
  message: ChatMessage;
  /** Hide sender label when this continues the previous agent turn (Cursor-like). */
  showSender?: boolean;
}

export const Message: React.FC<MessageProps> = ({ message, showSender = true }) => {
  const isUser = message.sender === 'user';

  if (message.kind === 'choice') {
    return (
      <div className="message-editorial message-choice">
        <WalkthroughAnswerChip answer={message.text} />
      </div>
    );
  }

  return (
    <div className={`message-editorial ${isUser ? 'message-user' : 'message-agent'}`}>
      {!isUser && showSender && <div className="message-sender">BlackJak</div>}
      <div className="message-body">
        <FileReferences text={message.text} />
      </div>
    </div>
  );
};
