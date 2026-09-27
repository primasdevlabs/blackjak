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
        <span className="card-title">Agent & Orchestration Settings</span>
      </div>

      <div className="form-group-list">
        <div className="form-row">
          <label>Effort Level:</label>
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
                <span>{lvl === 'extra_high' ? 'EXTRA HIGH' : lvl.toUpperCase()}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="form-row">
          <label>Operating Mode:</label>
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
                <span>{m === 'plan' ? 'PLAN MODE' : 'CODING MODE'}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="form-row">
          <label>Parallel Subagent Orchestration:</label>
          <label className="checkbox-row">
            <input
              type="checkbox"
              checked={settings.parallelSubagents}
              onChange={(e) => onUpdate({ parallelSubagents: e.target.checked })}
            />
            <span>Enable concurrent subagent execution</span>
          </label>
        </div>

        <div className="form-row">
          <label>Maximum Concurrent Subagents:</label>
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
          <label>Prompt Queue Execution Strategy:</label>
          <select
            className="select-input"
            value={settings.promptQueueBehavior}
            onChange={(e) => onUpdate({ promptQueueBehavior: e.target.value as any })}
          >
            <option value="sequential">Sequential (Default)</option>
            <option value="parallel">Parallel (Advanced)</option>
            <option value="ask">Ask before starting next</option>
          </select>
        </div>

        <div className="form-row">
          <label>Editor Preferences:</label>
          <div className="checkbox-group">
            <label className="checkbox-row">
              <input
                type="checkbox"
                checked={settings.autoOpenFile}
                onChange={(e) => onUpdate({ autoOpenFile: e.target.checked })}
              />
              <span>Automatically open newly created files</span>
            </label>

            <label className="checkbox-row">
              <input
                type="checkbox"
                checked={settings.autoOpenDiff}
                onChange={(e) => onUpdate({ autoOpenDiff: e.target.checked })}
              />
              <span>Automatically show native diffs on file edit</span>
            </label>

            <label className="checkbox-row">
              <input
                type="checkbox"
                checked={settings.askDestructiveOps}
                onChange={(e) => onUpdate({ askDestructiveOps: e.target.checked })}
              />
              <span>Require human approval for dangerous operations</span>
            </label>
          </div>
        </div>
      </div>
    </div>
  );
};
