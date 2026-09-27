import React from 'react';
import { SettingsConfig, GuardrailsConfig } from '../api/settings';

interface AgentSettingsProps {
  settings: SettingsConfig;
  onUpdate: (updates: Partial<SettingsConfig>) => void;
}

const DEFAULT_GUARDRAILS: GuardrailsConfig = {
  mode: 'supervised',
  shellAllowed: true,
  approveAllShell: false,
  denyCommands: [],
  protectedPaths: [],
  subagentsAllowed: true,
  maxSteps: 60,
};

export const AgentSettings: React.FC<AgentSettingsProps> = ({ settings, onUpdate }) => {
  const guardrails = settings.guardrails ?? DEFAULT_GUARDRAILS;
  const updateGuardrails = (patch: Partial<GuardrailsConfig>) => {
    onUpdate({ guardrails: { ...guardrails, ...patch } });
  };

  return (
    <div className="card agent-settings-card">
      <div className="card-header">
        <span className="card-title">Agent</span>
      </div>

      <div className="form-group-list">
        <div className="form-row">
          <label>Effort</label>
          <div className="radio-group-pills">
            {['low', 'medium', 'high', 'extra_high'].map((lvl) => (
              <label key={lvl} className={`pill-btn ${settings.effort === lvl ? 'active' : ''}`}>
                <input
                  type="radio"
                  name="effort"
                  value={lvl}
                  checked={settings.effort === lvl}
                  onChange={() => onUpdate({ effort: lvl as any })}
                />
                <span style={{ textTransform: 'capitalize' }}>{lvl.replace('_', ' ')}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="form-row">
          <label>Mode</label>
          <div className="radio-group-pills">
            {['plan', 'code'].map((m) => (
              <label key={m} className={`pill-btn ${settings.mode === m ? 'active' : ''}`}>
                <input
                  type="radio"
                  name="mode"
                  value={m}
                  checked={settings.mode === m}
                  onChange={() => onUpdate({ mode: m as any })}
                />
                <span style={{ textTransform: 'capitalize' }}>{m}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="form-row">
          <label>Subagents</label>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={settings.parallelSubagents}
              onChange={(e) => onUpdate({ parallelSubagents: e.target.checked })}
            />
            <span>Run subagents in parallel</span>
          </label>
        </div>

        <div className="form-row">
          <label>Max subagents</label>
          <input
            type="number"
            min="1"
            max="16"
            className="text-input num-input"
            value={settings.maxSubagents}
            onChange={(e) => onUpdate({ maxSubagents: parseInt(e.target.value, 10) || 4 })}
          />
        </div>

        <div className="form-row">
          <label>Prompt queue</label>
          <select
            className="select-input"
            value={settings.promptQueueBehavior}
            onChange={(e) => onUpdate({ promptQueueBehavior: e.target.value as any })}
          >
            <option value="sequential">Sequential</option>
            <option value="parallel">Parallel</option>
            <option value="ask">Ask first</option>
          </select>
        </div>

        <div className="form-row">
          <label>Editor</label>
          <div className="checkbox-group">
            <label className="checkbox-row">
              <input
                type="checkbox"
                checked={settings.autoOpenFile}
                onChange={(e) => onUpdate({ autoOpenFile: e.target.checked })}
              />
              <span>Open new files automatically</span>
            </label>

            <label className="checkbox-row">
              <input
                type="checkbox"
                checked={settings.autoOpenDiff}
                onChange={(e) => onUpdate({ autoOpenDiff: e.target.checked })}
              />
              <span>Show diffs on edit</span>
            </label>

            <label className="checkbox-row">
              <input
                type="checkbox"
                checked={settings.askDestructiveOps}
                onChange={(e) => onUpdate({ askDestructiveOps: e.target.checked })}
              />
              <span>Confirm destructive operations</span>
            </label>
          </div>
        </div>
      </div>

      <div className="card guardrails-card" style={{ marginTop: 16 }}>
        <div className="card-header">
          <span className="card-title">Guardrails</span>
        </div>

        <div className="form-group-list">
          <div className="form-row">
            <label>Autonomy</label>
            <div className="radio-group-pills">
              {(['supervised', 'readonly', 'autonomous'] as const).map((m) => (
                <label key={m} className={`pill-btn ${guardrails.mode === m ? 'active' : ''}`}>
                  <input
                    type="radio"
                    name="guardrails-mode"
                    value={m}
                    checked={guardrails.mode === m}
                    onChange={() => updateGuardrails({ mode: m })}
                  />
                  <span style={{ textTransform: 'capitalize' }}>
                    {m === 'readonly' ? 'Read-only' : m}
                  </span>
                </label>
              ))}
            </div>
            <p className="field-hint">
              Read-only blocks writes and unsafe commands. Autonomous skips all approval
              prompts — deny rules and protected paths still apply. Plan mode forces read-only.
            </p>
          </div>

          <div className="form-row">
            <label>Shell</label>
            <div className="checkbox-group">
              <label className="checkbox-row">
                <input
                  type="checkbox"
                  checked={guardrails.shellAllowed}
                  onChange={(e) => updateGuardrails({ shellAllowed: e.target.checked })}
                />
                <span>Allow shell commands</span>
              </label>
              <label className="checkbox-row">
                <input
                  type="checkbox"
                  checked={guardrails.approveAllShell}
                  onChange={(e) => updateGuardrails({ approveAllShell: e.target.checked })}
                />
                <span>Approve every shell command</span>
              </label>
            </div>
          </div>

          <div className="form-row">
            <label>Subagents</label>
            <label className="checkbox-row">
              <input
                type="checkbox"
                checked={guardrails.subagentsAllowed}
                onChange={(e) => updateGuardrails({ subagentsAllowed: e.target.checked })}
              />
              <span>Allow subagent delegation</span>
            </label>
          </div>

          <div className="form-row">
            <label>Max steps per run</label>
            <input
              type="number"
              min="1"
              max="500"
              className="text-input num-input"
              value={guardrails.maxSteps}
              onChange={(e) => updateGuardrails({ maxSteps: parseInt(e.target.value, 10) || 60 })}
            />
          </div>

          <div className="form-row form-row-column">
            <label>Blocked command patterns</label>
            <textarea
              className="text-input textarea-input"
              rows={3}
              placeholder={"One substring per line, e.g.\nnpm publish\ncurl | sh"}
              defaultValue={guardrails.denyCommands.join('\n')}
              onBlur={(e) =>
                updateGuardrails({
                  denyCommands: e.target.value.split('\n').map((s) => s.trim()).filter(Boolean),
                })
              }
            />
            <p className="field-hint">Commands containing any of these are always refused — no approval possible.</p>
          </div>

          <div className="form-row form-row-column">
            <label>Protected paths</label>
            <textarea
              className="text-input textarea-input"
              rows={3}
              placeholder={"Workspace-relative globs, e.g.\n.env.local\nsecrets/**"}
              defaultValue={guardrails.protectedPaths.join('\n')}
              onBlur={(e) =>
                updateGuardrails({
                  protectedPaths: e.target.value.split('\n').map((s) => s.trim()).filter(Boolean),
                })
              }
            />
            <p className="field-hint">
              The agent can never write these paths. Built-in rules (.git, .env, keys, .blackjak) always apply.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
