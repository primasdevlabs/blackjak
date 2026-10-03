import React, { useState } from 'react';

export interface WalkthroughOption {
  id: string;
  label: string;
}

interface WalkthroughQuestionProps {
  question: string;
  options: string[];
  allowOther?: boolean;
  onAnswer: (answer: string) => void;
  onDismiss?: () => void;
}

/** Cursor-style clickable walkthrough question with optional free-text Other. */
export const WalkthroughQuestion: React.FC<WalkthroughQuestionProps> = ({
  question,
  options,
  allowOther = true,
  onAnswer,
  onDismiss,
}) => {
  const [otherOpen, setOtherOpen] = useState(false);
  const [otherText, setOtherText] = useState('');

  const submitOther = () => {
    const t = otherText.trim();
    if (!t) return;
    onAnswer(t);
  };

  return (
    <div className="walkthrough-card">
      <div className="walkthrough-question">{question}</div>
      <div className="walkthrough-options">
        {options.map((opt, i) => (
          <button
            key={`${i}:${opt}`}
            type="button"
            className="walkthrough-option"
            onClick={() => onAnswer(opt)}
          >
            {opt}
          </button>
        ))}
        {allowOther && !otherOpen && (
          <button
            type="button"
            className="walkthrough-option walkthrough-other-toggle"
            onClick={() => setOtherOpen(true)}
          >
            Other…
          </button>
        )}
      </div>
      {allowOther && otherOpen && (
        <div className="walkthrough-other">
          <input
            className="walkthrough-other-input"
            autoFocus
            placeholder="Type your answer…"
            value={otherText}
            onChange={(e) => setOtherText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                submitOther();
              }
              if (e.key === 'Escape') {
                setOtherOpen(false);
              }
            }}
          />
          <button type="button" className="walkthrough-other-submit" onClick={submitOther} disabled={!otherText.trim()}>
            Submit
          </button>
        </div>
      )}
      {onDismiss && (
        <button type="button" className="walkthrough-dismiss" onClick={onDismiss}>
          Dismiss
        </button>
      )}
    </div>
  );
};

/** Compact chip shown in chat history after the user picks an option. */
export const WalkthroughAnswerChip: React.FC<{ answer: string }> = ({ answer }) => (
  <div className="walkthrough-answer-chip">{answer}</div>
);
