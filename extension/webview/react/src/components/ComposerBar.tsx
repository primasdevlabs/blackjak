import React, { useEffect, useRef, useState } from 'react';
import {
  ArrowUpIcon,
  ArrowsPointingInIcon,
  ChevronDownIcon,
  MicrophoneIcon,
  PaperClipIcon,
  PauseIcon,
  PlayIcon,
  StopIcon,
  SparklesIcon,
  ShieldCheckIcon,
  ChatBubbleLeftRightIcon,
} from '@heroicons/react/24/outline';
import { EffortLevel } from '../api/settings';
import { MentionPicker, MentionItem } from './MentionPicker';
import { agentHost } from '../host';
import { AgentModeUI, modeFromSettings } from '../utils/mode';

export type { AgentModeUI };
export { modeFromSettings };

interface ComposerBarProps {
  input: string;
  onInputChange: (value: string) => void;
  onSubmit: () => void;
  disabled: boolean;
  isWorking: boolean;
  isPaused: boolean;
  mode: AgentModeUI;
  effort: EffortLevel;
  modelLabel: string;
  models: { id: string; label: string; role: string }[];
  selectedModelId: string;
  onModeChange: (mode: AgentModeUI) => void;
  onEffortChange: (effort: EffortLevel) => void;
  onModelChange: (modelId: string, role?: string) => void;
  onAttach: (paths: string[]) => void;
  onPause: () => void;
  onStop: () => void;
  onResume: () => void;
  showSlashMenu: boolean;
  slashMenu: React.ReactNode;
  chips: React.ReactNode;
  compactChat: boolean;
  onToggleCompact: () => void;
}

const MODE_CYCLE: AgentModeUI[] = ['agent', 'plan', 'ask'];

export const ComposerBar: React.FC<ComposerBarProps> = ({
  input,
  onInputChange,
  onSubmit,
  disabled,
  isWorking,
  isPaused,
  mode,
  effort,
  modelLabel,
  models,
  selectedModelId,
  onModeChange,
  onEffortChange,
  onModelChange,
  onAttach,
  onPause,
  onStop,
  onResume,
  showSlashMenu,
  slashMenu,
  chips,
  compactChat,
  onToggleCompact,
}) => {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [showModeMenu, setShowModeMenu] = useState(false);
  const [showModelMenu, setShowModelMenu] = useState(false);
  const [mentionQuery, setMentionQuery] = useState<string | null>(null);
  const [listening, setListening] = useState(false);
  const recognitionRef = useRef<any>(null);

  const modeLabel = mode === 'ask' ? 'Ask' : mode === 'plan' ? 'Plan' : 'Agent';
  const ModeIcon = mode === 'ask' ? ChatBubbleLeftRightIcon : mode === 'plan' ? ShieldCheckIcon : SparklesIcon;
  const effortAccent = effort === 'high' || effort === 'extra_high';

  useEffect(() => {
    return () => {
      recognitionRef.current?.stop?.();
    };
  }, []);

  const detectMention = (value: string, caret: number) => {
    const before = value.slice(0, caret);
    const m = before.match(/@([^\s@]*)$/);
    if (m) setMentionQuery(m[1] ?? '');
    else setMentionQuery(null);
  };

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.target.value;
    onInputChange(val);
    detectMention(val, e.target.selectionStart ?? val.length);
  };

  const insertMention = (item: MentionItem) => {
    const el = textareaRef.current;
    const caret = el?.selectionStart ?? input.length;
    const before = input.slice(0, caret);
    const after = input.slice(caret);
    const token =
      item.type === 'symbol'
        ? `@${item.path}${item.line ? `:${item.line}` : ''}`
        : item.type === 'folder'
          ? `@folder:${item.path}`
          : `@${item.path}`;
    const replaced = before.replace(/@([^\s@]*)$/, `${token} `);
    onInputChange(replaced + after);
    setMentionQuery(null);
    // Also attach the file so it appears as a context chip (Cursor-like).
    if (item.type === 'file' || item.type === 'symbol') {
      onAttach([item.path]);
    }
    requestAnimationFrame(() => el?.focus());
  };

  const handleAttachClick = async () => {
    try {
      const paths = await agentHost.pickFiles({ canSelectMany: true, title: 'Attach files' });
      if (paths.length) onAttach(paths);
    } catch {
      const inputEl = document.createElement('input');
      inputEl.type = 'file';
      inputEl.multiple = true;
      inputEl.onchange = () => {
        const files = Array.from(inputEl.files ?? []);
        onAttach(files.map((f) => f.name));
      };
      inputEl.click();
    }
  };

  const toggleMic = () => {
    const SR = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
    if (!SR) {
      agentHost.showNotification('Speech recognition not available in this host', 'warn');
      return;
    }
    if (listening) {
      recognitionRef.current?.stop?.();
      setListening(false);
      return;
    }
    const rec = new SR();
    recognitionRef.current = rec;
    rec.continuous = false;
    rec.interimResults = true;
    rec.onresult = (ev: any) => {
      let transcript = '';
      for (let i = ev.resultIndex; i < ev.results.length; i++) {
        transcript += ev.results[i][0].transcript;
      }
      if (transcript) onInputChange((input ? input + ' ' : '') + transcript.trim());
    };
    rec.onend = () => setListening(false);
    rec.onerror = () => setListening(false);
    setListening(true);
    rec.start();
  };

  const cycleMode = () => {
    const idx = MODE_CYCLE.indexOf(mode);
    onModeChange(MODE_CYCLE[(idx + 1) % MODE_CYCLE.length]);
  };

  return (
    <div className="composer-card composer-bar">
      {chips}
      {showSlashMenu && slashMenu}
      {mentionQuery !== null && (
        <MentionPicker
          query={mentionQuery}
          onSelect={insertMention}
          onClose={() => setMentionQuery(null)}
        />
      )}

      <textarea
        ref={textareaRef}
        className="composer-input"
        placeholder={
          disabled
            ? 'Connecting…'
            : isWorking
              ? 'Add a follow-up'
              : 'Ask a question or describe a task. Use / for commands, @ for context'
        }
        value={input}
        onChange={handleChange}
        onKeyDown={(e) => {
          if (mentionQuery !== null && (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter' || e.key === 'Escape')) {
            return;
          }
          if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            onSubmit();
          }
        }}
        disabled={disabled}
        rows={3}
      />

      <div className="composer-toolbar">
        <div className="composer-actions">
          <button className="composer-icon-btn" title="Attach files" onClick={handleAttachClick}>
            <PaperClipIcon className="icon-sm" />
          </button>

          <div className="composer-anchored">
            <button
              className={`chip-btn mode-pill mode-${mode}`}
              onClick={() => setShowModeMenu(!showModeMenu)}
              onContextMenu={(e) => {
                e.preventDefault();
                cycleMode();
              }}
              title="Agent mode — click to change"
            >
              <ModeIcon className="icon-sm" />
              <span>{modeLabel}</span>
              <ChevronDownIcon className="icon-xs" />
            </button>
            {showModeMenu && (
              <>
                <div className="dropdown-backdrop" onClick={() => setShowModeMenu(false)} />
                <div className="composer-dropdown">
                  {MODE_CYCLE.map((m) => (
                    <button
                      key={m}
                      className={`composer-dropdown-item ${mode === m ? 'active' : ''}`}
                      onClick={() => {
                        onModeChange(m);
                        setShowModeMenu(false);
                      }}
                    >
                      <span style={{ textTransform: 'capitalize' }}>{m}</span>
                      {mode === m && <span className="dropdown-check">✓</span>}
                    </button>
                  ))}
                </div>
              </>
            )}
          </div>

          <div className="composer-anchored">
            <button
              className="chip-btn model-pill"
              onClick={() => setShowModelMenu(!showModelMenu)}
              title="Model & effort"
            >
              <span>{modelLabel}</span>
              <span className={effortAccent ? 'effort-accent' : ''} style={{ textTransform: 'capitalize' }}>
                {effort.replace('_', ' ')}
              </span>
              <ChevronDownIcon className="icon-xs" />
            </button>
            {showModelMenu && (
              <>
                <div className="dropdown-backdrop" onClick={() => setShowModelMenu(false)} />
                <div className="composer-dropdown model-menu">
                  <div className="composer-dropdown-section">Model</div>
                  {models.map((m) => (
                    <button
                      key={m.id}
                      className={`composer-dropdown-item ${selectedModelId === m.id ? 'active' : ''}`}
                      onClick={() => {
                        onModelChange(m.id, m.role);
                        setShowModelMenu(false);
                      }}
                    >
                      <span>{m.label}</span>
                      <span className="dropdown-meta">{m.role}</span>
                    </button>
                  ))}
                  <div className="composer-dropdown-divider" />
                  <div className="composer-dropdown-section">Effort</div>
                  {(['low', 'medium', 'high', 'extra_high'] as const).map((level) => (
                    <button
                      key={level}
                      className={`composer-dropdown-item ${effort === level ? 'active' : ''}`}
                      onClick={() => {
                        onEffortChange(level);
                        setShowModelMenu(false);
                      }}
                    >
                      <span style={{ textTransform: 'capitalize' }}>{level.replace('_', ' ')}</span>
                      {effort === level && <span className="dropdown-check">✓</span>}
                    </button>
                  ))}
                </div>
              </>
            )}
          </div>

          <button
            className={`composer-icon-btn ${compactChat ? 'active' : ''}`}
            title="Compact chat"
            onClick={onToggleCompact}
          >
            <ArrowsPointingInIcon className="icon-sm" />
          </button>
        </div>

        <div className="composer-actions">
          <button
            className={`composer-icon-btn mic-btn ${listening ? 'listening' : ''}`}
            title="Voice input"
            onClick={toggleMic}
          >
            <MicrophoneIcon className="icon-sm" />
          </button>
          {isWorking ? (
            <>
              <button className="composer-send-btn" onClick={onPause} title="Pause">
                <PauseIcon className="icon-sm" />
              </button>
              <button className="composer-send-btn composer-send-btn-stop" onClick={onStop} title="Stop">
                <StopIcon className="icon-sm" />
              </button>
            </>
          ) : isPaused ? (
            <button className="composer-send-btn" onClick={onResume} title="Resume">
              <PlayIcon className="icon-sm" />
            </button>
          ) : (
            <button
              className="composer-send-btn"
              onClick={onSubmit}
              disabled={disabled || !input.trim()}
              title="Send"
            >
              <ArrowUpIcon className="icon-sm" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
