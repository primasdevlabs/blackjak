import React from 'react';
import { TestResult } from '../state/agentStore';
import { BeakerIcon } from '@heroicons/react/24/outline';

interface TestResultsProps {
  results: TestResult[];
}

export const TestResults: React.FC<TestResultsProps> = ({ results }) => {
  if (results.length === 0) return null;

  return (
    <div className="card test-card">
      <div className="card-header">
        <BeakerIcon className="icon header-icon" />
        <span className="card-title">Test Results</span>
      </div>
      <div className="test-list">
        {results.map((r, idx) => (
          <div key={idx} className="test-item">
            <div className="test-summary">
              <span className="suite-name">{r.suite}</span>
              <span className="test-badge passed">{r.passed} Passed</span>
              {r.failed > 0 && <span className="test-badge failed">{r.failed} Failed</span>}
            </div>
            {r.output && <pre className="test-output">{r.output}</pre>}
          </div>
        ))}
      </div>
    </div>
  );
};
