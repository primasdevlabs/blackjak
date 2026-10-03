import React, { useEffect, useState } from 'react';
import { apiClient } from '../api/client';
import { SettingsConfig } from '../api/settings';

interface RuleItem {
  id: string;
  name: string;
  body: string;
  source: string;
  enabled: boolean;
}

interface SkillItem {
  name: string;
  description: string;
  enabled: boolean;
  path: string;
}

export const GeneralSettings: React.FC<{
  settings: SettingsConfig;
  onUpdate: (u: Partial<SettingsConfig>) => void;
}> = ({ settings, onUpdate }) => (
  <div className="card">
    <div className="card-header"><span className="card-title">General</span></div>
    <div className="form-group-list">
      <div className="form-row">
        <label>Default mode</label>
        <select
          value={settings.mode === 'code' ? 'agent' : settings.mode}
          onChange={(e) => onUpdate({ mode: e.target.value as any })}
        >
          <option value="agent">Agent</option>
          <option value="plan">Plan</option>
          <option value="ask">Ask</option>
        </select>
      </div>
      <div className="form-row">
        <label>Default effort</label>
        <select value={settings.effort} onChange={(e) => onUpdate({ effort: e.target.value as any })}>
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
          <option value="extra_high">Extra high</option>
        </select>
      </div>
      <div className="form-row">
        <label>Density</label>
        <select
          value={settings.themeDensity || 'comfortable'}
          onChange={(e) => onUpdate({ themeDensity: e.target.value as any })}
        >
          <option value="comfortable">Comfortable</option>
          <option value="compact">Compact</option>
        </select>
      </div>
      <label className="checkbox-row">
        <input
          type="checkbox"
          checked={!!settings.compactChatDefault}
          onChange={(e) => onUpdate({ compactChatDefault: e.target.checked })}
        />
        <span>Compact chat by default</span>
      </label>
    </div>
  </div>
);

export const UsageSettings: React.FC = () => {
  const [usage, setUsage] = useState<any>(null);
  useEffect(() => {
    fetch(`${apiClient.getBaseUrl()}/api/usage`)
      .then((r) => r.json())
      .then((j) => j.success && setUsage(j.data))
      .catch(() => undefined);
  }, []);
  return (
    <div className="card">
      <div className="card-header"><span className="card-title">Plan &amp; Usage</span></div>
      <p className="settings-hint">Local token and cache estimates — no billing.</p>
      <div className="usage-grid">
        <div><span>Runs</span><strong>{usage?.runs ?? 0}</strong></div>
        <div><span>Input tokens</span><strong>{usage?.inputTokens ?? 0}</strong></div>
        <div><span>Output tokens</span><strong>{usage?.outputTokens ?? 0}</strong></div>
        <div><span>Cache read</span><strong>{usage?.cacheReadTokens ?? 0}</strong></div>
        <div><span>Cache write</span><strong>{usage?.cacheWriteTokens ?? 0}</strong></div>
        <div><span>Cache hit rate</span><strong>{Math.round((usage?.cacheHitRate ?? 0) * 100)}%</strong></div>
      </div>
    </div>
  );
};

export const RulesSettings: React.FC = () => {
  const [rules, setRules] = useState<RuleItem[]>([]);
  const [name, setName] = useState('');
  const [body, setBody] = useState('');
  const [source, setSource] = useState<'project' | 'user'>('project');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [draftBody, setDraftBody] = useState('');
  const [saving, setSaving] = useState(false);

  const load = () => {
    fetch(`${apiClient.getBaseUrl()}/api/rules`)
      .then((r) => r.json())
      .then((j) => j.success && setRules(j.data || []))
      .catch(() => setRules([]));
  };
  useEffect(() => { load(); }, []);

  const addRule = async () => {
    if (!name.trim() || !body.trim()) return;
    setSaving(true);
    try {
      await fetch(`${apiClient.getBaseUrl()}/api/rules`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), body, enabled: true, source }),
      });
      setName('');
      setBody('');
      load();
    } finally {
      setSaving(false);
    }
  };

  const toggle = async (id: string, enabled: boolean) => {
    await fetch(`${apiClient.getBaseUrl()}/api/rules/enable`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, enabled }),
    });
    load();
  };

  const startEdit = (r: RuleItem) => {
    setEditingId(r.id);
    setDraftBody(r.body);
  };

  const saveEdit = async (r: RuleItem) => {
    setSaving(true);
    try {
      await fetch(`${apiClient.getBaseUrl()}/api/rules/update`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: r.id,
          name: r.name,
          body: draftBody,
          source: r.source === 'user' ? 'user' : 'project',
          enabled: r.enabled,
        }),
      });
      setEditingId(null);
      load();
    } finally {
      setSaving(false);
    }
  };

  const remove = async (id: string) => {
    await fetch(`${apiClient.getBaseUrl()}/api/rules/delete`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id }),
    });
    if (editingId === id) setEditingId(null);
    load();
  };

  return (
    <div className="card">
      <div className="card-header"><span className="card-title">Rules &amp; Policies</span></div>
      <p className="settings-hint">
        Project rules live in <code>.blackjak/rules/*.md</code>. Editing a builtin saves a project override.
        Disabled rules are excluded from the agent prompt.
      </p>

      <div className="form-group-list rules-add-form">
        <div className="form-row">
          <label>New rule</label>
          <div className="radio-group-pills">
            <label className={`pill-btn ${source === 'project' ? 'active' : ''}`}>
              <input type="radio" name="rule-source" checked={source === 'project'} onChange={() => setSource('project')} />
              <span>Project</span>
            </label>
            <label className={`pill-btn ${source === 'user' ? 'active' : ''}`}>
              <input type="radio" name="rule-source" checked={source === 'user'} onChange={() => setSource('user')} />
              <span>User</span>
            </label>
          </div>
        </div>
        <input
          className="text-input"
          placeholder="Rule name (e.g. api-conventions)"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <textarea
          className="text-input textarea-input"
          placeholder="Rule body"
          rows={4}
          value={body}
          onChange={(e) => setBody(e.target.value)}
        />
        <button type="button" className="btn btn-primary" onClick={addRule} disabled={saving || !name.trim() || !body.trim()}>
          Add rule
        </button>
      </div>

      {rules.length === 0 && <div className="settings-hint">No rules found yet.</div>}

      <div className="rules-list">
        {rules.map((r) => {
          const editing = editingId === r.id;
          const canDelete = r.source === 'user' || r.source === 'project';
          return (
            <div key={r.id} className={`rules-item ${r.enabled ? '' : 'rules-item-disabled'}`}>
              <div className="rules-item-header">
                <label className="checkbox-row rules-item-toggle">
                  <input
                    type="checkbox"
                    checked={r.enabled}
                    onChange={(e) => void toggle(r.id, e.target.checked)}
                  />
                  <span className="rules-item-name">{r.name}</span>
                </label>
                <span className={`rules-source-badge source-${r.source}`}>{r.source}</span>
              </div>

              {editing ? (
                <div className="rules-item-editor">
                  <textarea
                    className="text-input textarea-input"
                    rows={8}
                    value={draftBody}
                    onChange={(e) => setDraftBody(e.target.value)}
                  />
                  <div className="rules-item-actions">
                    <button type="button" className="btn btn-primary" disabled={saving} onClick={() => void saveEdit(r)}>
                      Save
                    </button>
                    <button type="button" className="btn btn-secondary" onClick={() => setEditingId(null)}>
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <>
                  <pre className="rules-item-preview">{r.body.slice(0, 280)}{r.body.length > 280 ? '…' : ''}</pre>
                  <div className="rules-item-actions">
                    <button type="button" className="chip-btn" onClick={() => startEdit(r)}>
                      Edit
                    </button>
                    {canDelete && (
                      <button type="button" className="chip-btn" onClick={() => void remove(r.id)}>
                        Delete
                      </button>
                    )}
                  </div>
                </>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export const SkillsSettings: React.FC = () => {
  const [skills, setSkills] = useState<SkillItem[]>([]);
  const load = () => {
    fetch(`${apiClient.getBaseUrl()}/api/skills`)
      .then((r) => r.json())
      .then((j) => j.success && setSkills(j.data || []))
      .catch(() => setSkills([]));
  };
  useEffect(() => { load(); }, []);

  const toggle = async (name: string, enabled: boolean) => {
    await fetch(`${apiClient.getBaseUrl()}/api/skills/${encodeURIComponent(name)}/enable`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled }),
    });
    load();
  };

  return (
    <div className="card">
      <div className="card-header"><span className="card-title">Skills</span></div>
      <p className="settings-hint">Install packs under <code>.blackjak/skills/&lt;name&gt;/SKILL.md</code> and invoke with <code>/</code>.</p>
      {skills.length === 0 && <div className="settings-hint">No skills found yet.</div>}
      {skills.map((s) => (
        <label key={s.name} className="checkbox-row skills-row">
          <input type="checkbox" checked={s.enabled} onChange={(e) => toggle(s.name, e.target.checked)} />
          <span>
            <strong>{s.name}</strong>
            <div className="settings-hint">{s.description}</div>
          </span>
        </label>
      ))}
    </div>
  );
};

export const IndexingSettings: React.FC = () => {
  const [stats, setStats] = useState<any>(null);
  const load = () => {
    fetch(`${apiClient.getBaseUrl()}/api/index`)
      .then((r) => r.json())
      .then((j) => j.success && setStats(j.data))
      .catch(() => undefined);
  };
  useEffect(() => { load(); }, []);
  const rebuild = async () => {
    await fetch(`${apiClient.getBaseUrl()}/api/index/rebuild`, { method: 'POST' });
    setTimeout(load, 500);
  };
  return (
    <div className="card">
      <div className="card-header"><span className="card-title">Indexing</span></div>
      <div className="usage-grid">
        <div><span>Status</span><strong>{stats?.status ?? '—'}</strong></div>
        <div><span>Files</span><strong>{stats?.fileCount ?? 0}</strong></div>
        <div><span>Last indexed</span><strong>{stats?.lastIndexed ? new Date(stats.lastIndexed).toLocaleString() : '—'}</strong></div>
      </div>
      <button type="button" className="btn btn-primary" onClick={rebuild} style={{ marginTop: 12 }}>Rebuild index</button>
    </div>
  );
};

export const CustomizeSettings: React.FC<{
  settings: SettingsConfig;
  onUpdate: (u: Partial<SettingsConfig>) => void;
}> = ({ settings, onUpdate }) => (
  <div className="card">
    <div className="card-header"><span className="card-title">Customize</span></div>
    <label className="checkbox-row">
      <input
        type="checkbox"
        checked={!!settings.compactChatDefault}
        onChange={(e) => onUpdate({ compactChatDefault: e.target.checked })}
      />
      <span>Compact chat default</span>
    </label>
    <div className="form-row" style={{ marginTop: 12 }}>
      <label>Theme density</label>
      <select
        value={settings.themeDensity || 'comfortable'}
        onChange={(e) => onUpdate({ themeDensity: e.target.value as any })}
      >
        <option value="comfortable">Comfortable</option>
        <option value="compact">Compact</option>
      </select>
    </div>
  </div>
);

export const GitPRSettings: React.FC<{
  settings: SettingsConfig;
  onUpdate: (u: Partial<SettingsConfig>) => void;
}> = ({ settings, onUpdate }) => (
  <div className="card">
    <div className="card-header"><span className="card-title">Git &amp; PRs</span></div>
    <label className="checkbox-row">
      <input type="checkbox" checked={settings.autoOpenDiff} onChange={(e) => onUpdate({ autoOpenDiff: e.target.checked })} />
      <span>Auto-open diffs</span>
    </label>
    <label className="checkbox-row">
      <input type="checkbox" checked={settings.autoOpenFile} onChange={(e) => onUpdate({ autoOpenFile: e.target.checked })} />
      <span>Auto-open touched files</span>
    </label>
    <label className="checkbox-row">
      <input type="checkbox" checked={settings.askDestructiveOps} onChange={(e) => onUpdate({ askDestructiveOps: e.target.checked })} />
      <span>Approve destructive git ops</span>
    </label>
  </div>
);

export const BrowserNetworkSettings: React.FC<{
  settings: SettingsConfig;
  onUpdate: (u: Partial<SettingsConfig>) => void;
}> = ({ settings, onUpdate }) => (
  <div className="card">
    <div className="card-header"><span className="card-title">Browser &amp; Network</span></div>
    <div className="form-row">
      <label>Proxy URL</label>
      <input
        value={settings.proxyUrl || ''}
        onChange={(e) => onUpdate({ proxyUrl: e.target.value })}
        placeholder="http://127.0.0.1:7890"
      />
    </div>
    <div className="form-row">
      <label>Allowlist (comma-separated hosts)</label>
      <input
        value={(settings.networkAllowlist || []).join(', ')}
        onChange={(e) =>
          onUpdate({
            networkAllowlist: e.target.value.split(',').map((s) => s.trim()).filter(Boolean),
          })
        }
      />
    </div>
  </div>
);

export const TabSettingsPanel: React.FC<{
  settings: SettingsConfig;
  onUpdate: (u: Partial<SettingsConfig>) => void;
}> = ({ settings, onUpdate }) => (
  <div className="card">
    <div className="card-header"><span className="card-title">Tab</span></div>
    <div className="form-row">
      <label>Auto-open file limit</label>
      <input
        type="number"
        min={1}
        max={20}
        value={settings.tabAutoOpenLimit ?? 5}
        onChange={(e) => onUpdate({ tabAutoOpenLimit: Number(e.target.value) || 5 })}
      />
    </div>
  </div>
);

export const BetaSettings: React.FC<{
  settings: SettingsConfig;
  onUpdate: (u: Partial<SettingsConfig>) => void;
}> = ({ settings, onUpdate }) => {
  const flags = settings.betaFlags || {};
  const toggle = (key: string) =>
    onUpdate({ betaFlags: { ...flags, [key]: !flags[key] } });
  return (
    <div className="card">
      <div className="card-header"><span className="card-title">Beta</span></div>
      <label className="checkbox-row">
        <input type="checkbox" checked={!!flags.browserTool} onChange={() => toggle('browserTool')} />
        <span>Browser tool (allowlisted fetch)</span>
      </label>
      <label className="checkbox-row">
        <input type="checkbox" checked={!!flags.promptCache} onChange={() => toggle('promptCache')} />
        <span>Prompt caching</span>
      </label>
    </div>
  );
};
