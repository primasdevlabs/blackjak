import React from 'react';

interface ModelRoutingSettingsProps {
  routes: Record<string, string>;
  onChangeRoute: (task: string, role: string) => void;
  onResetDefaults: () => void;
}

export const ModelRoutingSettings: React.FC<ModelRoutingSettingsProps> = ({
  routes,
  onChangeRoute,
  onResetDefaults,
}) => {
  const taskTypes = [
    { id: 'planning', label: 'Planning & Architecture' },
    { id: 'exploration', label: 'Exploration & Search' },
    { id: 'coding', label: 'Coding & Refactoring' },
    { id: 'debugging', label: 'Debugging Strategy' },
    { id: 'testing', label: 'Test Execution' },
    { id: 'review', label: 'Code Review' },
    { id: 'summarization', label: 'Summarization' },
  ];

  return (
    <div className="card routing-settings-card">
      <div className="card-header">
        <span className="card-title">Task Model Routing</span>
        <button className="btn btn-sm btn-deny reset-btn" onClick={onResetDefaults}>
          Reset Defaults
        </button>
      </div>

      <div className="routing-table">
        {taskTypes.map((task) => (
          <div key={task.id} className="routing-row">
            <span className="task-label">{task.label}:</span>
            <select
              className="select-input role-select"
              value={routes[task.id] || 'coding'}
              onChange={(e) => onChangeRoute(task.id, e.target.value)}
            >
              <option value="thinking">Thinking Model</option>
              <option value="coding">Coding Model</option>
              <option value="fast">Fast Model</option>
            </select>
          </div>
        ))}
      </div>
    </div>
  );
};
