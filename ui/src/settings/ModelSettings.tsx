import React from 'react';
import { Model } from '../api/settings';

interface ModelSettingsProps {
  useSeparateModels: boolean;
  thinkingModelId: string;
  codingModelId: string;
  fastModelId: string;
  availableModels: Model[];
  onChangeSeparate: (val: boolean) => void;
  onChangeThinking: (id: string) => void;
  onChangeCoding: (id: string) => void;
  onChangeFast: (id: string) => void;
}

export const ModelSettings: React.FC<ModelSettingsProps> = ({
  useSeparateModels,
  thinkingModelId,
  codingModelId,
  fastModelId,
  availableModels,
  onChangeSeparate,
  onChangeThinking,
  onChangeCoding,
  onChangeFast,
}) => {
  return (
    <div className="card model-settings-card">
      <div className="card-header">
        <span className="card-title">Model Role Selection</span>
      </div>

      <div className="form-group-list">
        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={useSeparateModels}
            onChange={(e) => onChangeSeparate(e.target.checked)}
          />
          <span>Use separate Thinking and Coding models</span>
        </label>

        {useSeparateModels ? (
          <>
            <div className="form-row">
              <label>Thinking / Reasoning Model:</label>
              <select
                className="select-input"
                value={thinkingModelId}
                onChange={(e) => onChangeThinking(e.target.value)}
              >
                {availableModels.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name} ({m.provider}) {m.supportsReasoning ? '[Reasoning]' : ''}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-row">
              <label>Coding / Execution Model:</label>
              <select
                className="select-input"
                value={codingModelId}
                onChange={(e) => onChangeCoding(e.target.value)}
              >
                {availableModels.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name} ({m.provider}) {m.supportsTools ? '[Tools]' : ''}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-row">
              <label>Fast / Utility Model:</label>
              <select
                className="select-input"
                value={fastModelId}
                onChange={(e) => onChangeFast(e.target.value)}
              >
                {availableModels.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.name} ({m.provider})
                  </option>
                ))}
              </select>
            </div>
          </>
        ) : (
          <div className="form-row">
            <label>Primary Unified Model:</label>
            <select
              className="select-input"
              value={codingModelId}
              onChange={(e) => onChangeCoding(e.target.value)}
            >
              {availableModels.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.name} ({m.provider})
                </option>
              ))}
            </select>
          </div>
        )}
      </div>
    </div>
  );
};
