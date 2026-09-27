import React from 'react';
import { SettingsConfig } from '../api/settings';

interface AgentSettingsProps {
  settings: SettingsConfig;
  onUpdate: (updates: Partial<SettingsConfig>) => void;
}

export const AgentSettings: React.FC<AgentSettingsProps> = ({ settings, onUpdate }) => {
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
    </div>
  );
};
