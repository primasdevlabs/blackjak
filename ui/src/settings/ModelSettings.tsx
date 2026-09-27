import React, { useState } from 'react';
import { Model, EffortLevel, ModelRole, ModelConfig } from '../api/settings';
import { settingsStore } from '../state/settingsStore';
import { ArrowPathIcon, InformationCircleIcon } from '@heroicons/react/24/outline';

interface ModelSettingsProps {
  useSeparateModels: boolean;
  thinking: ModelConfig;
  coding: ModelConfig;
  fast: ModelConfig;
  review: ModelConfig;
  effort: EffortLevel;
  availableModelsMap: Record<string, Model[]>;
  onChangeSeparate: (val: boolean) => void;
  onChangeModelConfig: (role: ModelRole, config: ModelConfig) => void;
  onChangeEffort: (effort: EffortLevel) => void;
}

const PROVIDER_OPTIONS = ['OpenAI', 'Google Gemini', 'Anthropic', 'OpenAI-compatible'];

export const ModelSettings: React.FC<ModelSettingsProps> = ({
  useSeparateModels,
  thinking,
  coding,
  fast,
  review,
  effort,
  availableModelsMap,
  onChangeSeparate,
  onChangeModelConfig,
  onChangeEffort,
}) => {
  const [refreshing, setRefreshing] = useState(false);

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await settingsStore.refreshModels();
    } catch (err) {
      console.error('Failed to refresh model catalog:', err);
    } finally {
      setRefreshing(false);
    }
  };

  const getModelsForRole = (providerId: string, role: ModelRole): Model[] => {
    const providerModels = availableModelsMap[providerId] || [];
    if (providerModels.length === 0) {
      // Return catalog defaults if dynamic list empty
      return getFallbackModels(providerId, role);
    }
    // Filter out non-coding categories (audio, image, embedding)
    const filtered = providerModels.filter(m => !m.category || ['coding', 'reasoning', 'fast', 'openweight'].includes(m.category));
    return filtered.length > 0 ? filtered : providerModels;
  };

  const getFallbackModels = (providerId: string, role: ModelRole): Model[] => {
    switch (providerId) {
      case 'OpenAI':
        if (role === 'thinking') return [{ id: 'gpt-6-astra', name: 'GPT-6 Astra', provider: 'OpenAI', supportsTools: true, supportsVision: true, supportsReasoning: true, supportsEffort: true, contextWindow: 500000, defaultMaxTokens: 32768 }];
        if (role === 'coding') return [{ id: 'gpt-5.3-codex', name: 'GPT-5.3-Codex', provider: 'OpenAI', supportsTools: true, supportsVision: true, supportsReasoning: true, supportsEffort: true, contextWindow: 256000, defaultMaxTokens: 16384 }];
        if (role === 'fast') return [{ id: 'gpt-5.4-mini', name: 'GPT-5.4 Mini', provider: 'OpenAI', supportsTools: true, supportsVision: true, supportsReasoning: false, supportsEffort: false, contextWindow: 128000, defaultMaxTokens: 4096 }];
        return [{ id: 'gpt-4o', name: 'GPT-4o', provider: 'OpenAI', supportsTools: true, supportsVision: true, supportsReasoning: false, supportsEffort: false, contextWindow: 128000, defaultMaxTokens: 4096 }];
      case 'Google Gemini':
        if (role === 'thinking') return [{ id: 'gemini-3.1-pro-preview', name: 'Gemini 3.1 Pro Preview', provider: 'Google Gemini', supportsTools: true, supportsVision: true, supportsReasoning: true, supportsEffort: true, contextWindow: 2000000, defaultMaxTokens: 16384 }];
        if (role === 'coding') return [{ id: 'gemini-3.8-flash', name: 'Gemini 3.8 Flash', provider: 'Google Gemini', supportsTools: true, supportsVision: true, supportsReasoning: true, supportsEffort: true, contextWindow: 2000000, defaultMaxTokens: 16384 }];
        return [{ id: 'gemini-3.5-flash-lite', name: 'Gemini 3.5 Flash-Lite', provider: 'Google Gemini', supportsTools: true, supportsVision: true, supportsReasoning: false, supportsEffort: false, contextWindow: 1000000, defaultMaxTokens: 8192 }];
      case 'Anthropic':
        if (role === 'thinking') return [{ id: 'claude-opus-5', name: 'Claude Opus 5', provider: 'Anthropic', supportsTools: true, supportsVision: true, supportsReasoning: true, supportsEffort: true, contextWindow: 300000, defaultMaxTokens: 16384 }];
        if (role === 'review') return [{ id: 'claude-sonnet-5', name: 'Claude Sonnet 5', provider: 'Anthropic', supportsTools: true, supportsVision: true, supportsReasoning: true, supportsEffort: true, contextWindow: 300000, defaultMaxTokens: 16384 }];
        if (role === 'fast') return [{ id: 'claude-haiku-4.5', name: 'Claude Haiku 4.5', provider: 'Anthropic', supportsTools: true, supportsVision: true, supportsReasoning: false, supportsEffort: false, contextWindow: 200000, defaultMaxTokens: 8192 }];
        return [{ id: 'claude-sonnet-4.6', name: 'Claude Sonnet 4.6', provider: 'Anthropic', supportsTools: true, supportsVision: true, supportsReasoning: true, supportsEffort: true, contextWindow: 200000, defaultMaxTokens: 8192 }];
      default:
        return [{ id: 'llama3.2', name: 'llama3.2 (Compatible)', provider: 'OpenAI-compatible', supportsTools: true, supportsVision: false, supportsReasoning: false, supportsEffort: false, contextWindow: 32768, defaultMaxTokens: 4096 }];
    }
  };

  const selectedCodingModel = (availableModelsMap[coding.providerId] || []).find(m => m.id === coding.modelId);
  const supportsNativeEffort = selectedCodingModel ? selectedCodingModel.supportsEffort : true;

  const renderRoleSelector = (label: string, role: ModelRole, config: ModelConfig) => {
    const models = getModelsForRole(config.providerId, role);
    return (
      <div className="form-group" style={{ marginBottom: '16px' }}>
        <label style={{ display: 'block', fontWeight: 600, marginBottom: '6px' }}>{label}</label>
        <div style={{ display: 'grid', gridTemplateColumns: '140px 1fr', gap: '8px' }}>
          <select
            className="select-input"
            value={config.providerId}
            onChange={(e) => {
              const newProv = e.target.value;
              const newModels = getModelsForRole(newProv, role);
              const firstModelId = newModels.length > 0 ? newModels[0].id : '';
              onChangeModelConfig(role, { providerId: newProv, modelId: firstModelId, role });
            }}
          >
            {PROVIDER_OPTIONS.map((p) => (
              <option key={p} value={p}>{p}</option>
            ))}
          </select>

          <select
            className="select-input"
            value={config.modelId}
            onChange={(e) => onChangeModelConfig(role, { ...config, modelId: e.target.value })}
          >
            {models.map((m) => (
              <option key={m.id} value={m.id}>
                {m.name} {m.supportsReasoning ? '[Reasoning]' : ''} {m.supportsEffort ? '[Effort Control]' : ''}
              </option>
            ))}
          </select>
        </div>
      </div>
    );
  };

  return (
    <div className="card model-settings-card">
      <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span className="card-title">MODELS</span>
        <button
          className="btn btn-secondary btn-sm"
          onClick={handleRefresh}
          disabled={refreshing}
          style={{ display: 'flex', alignItems: 'center', gap: '4px' }}
        >
          <ArrowPathIcon className={`icon-sm ${refreshing ? 'animate-spin' : ''}`} />
          {refreshing ? 'Refreshing...' : 'Refresh Models'}
        </button>
      </div>

      <div className="form-group-list" style={{ marginTop: '16px' }}>
        <label className="checkbox-row" style={{ marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
          <input
            type="checkbox"
            checked={useSeparateModels}
            onChange={(e) => onChangeSeparate(e.target.checked)}
          />
          <span>Use separate thinking and coding models</span>
        </label>

        {useSeparateModels ? (
          <>
            {renderRoleSelector('Thinking Model', 'thinking', thinking)}
            {renderRoleSelector('Coding Model', 'coding', coding)}
            {renderRoleSelector('Fast Model', 'fast', fast)}
            {renderRoleSelector('Review Model', 'review', review)}
          </>
        ) : (
          renderRoleSelector('Primary Model', 'coding', coding)
        )}

        <div className="form-group" style={{ marginTop: '24px' }}>
          <label style={{ display: 'block', fontWeight: 600, marginBottom: '8px' }}>Effort Level</label>
          <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
            {(['low', 'medium', 'high', 'extra_high'] as EffortLevel[]).map((level) => (
              <label key={level} style={{ display: 'flex', alignItems: 'center', gap: '6px', cursor: 'pointer' }}>
                <input
                  type="radio"
                  name="effortLevel"
                  value={level}
                  checked={effort === level}
                  onChange={() => onChangeEffort(level)}
                />
                <span style={{ textTransform: 'capitalize' }}>
                  {level === 'extra_high' ? 'Extra High' : level}
                </span>
              </label>
            ))}
          </div>

          {!supportsNativeEffort && (
            <div className="info-banner" style={{ marginTop: '12px', display: 'flex', alignItems: 'flex-start', gap: '8px', padding: '10px 12px', background: 'var(--bg-card-hover)', borderRadius: '6px', borderLeft: '3px solid var(--accent-color)', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
              <InformationCircleIcon className="icon-sm" style={{ flexShrink: 0, marginTop: '2px' }} />
              <span>
                This provider/model does not expose native effort control. Agent orchestration will use effort to adjust planning depth, subagent count, and review passes.
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
