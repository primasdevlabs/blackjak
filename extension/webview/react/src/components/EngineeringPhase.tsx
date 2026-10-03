import React from 'react';

export interface EngineeringEvidence {
  inspect?: boolean;
  plan?: boolean;
  implement?: boolean;
  verify?: boolean;
  review?: boolean;
}

export interface EngineeringPhaseState {
  phase: string;
  taskClass: string;
  evidence: EngineeringEvidence;
  filesChanged?: boolean;
  review?: {
    verdict?: string;
    findings?: string[];
    solvesProblem?: boolean;
    unnecessaryComplexity?: boolean;
    aiSlopRisk?: boolean;
    securityIssues?: string[];
  };
}

interface Props {
  engineering: EngineeringPhaseState | null;
}

const EVIDENCE_KEYS: { key: keyof EngineeringEvidence; label: string }[] = [
  { key: 'inspect', label: 'Inspect' },
  { key: 'plan', label: 'Plan' },
  { key: 'verify', label: 'Verify' },
  { key: 'review', label: 'Review' },
];

export const EngineeringPhaseStrip: React.FC<Props> = ({ engineering }) => {
  // Keep the main chat quiet for simple Q&A — only show for full engineering tasks.
  if (!engineering || engineering.taskClass === 'trivial') return null;

  return (
    <div className="eng-phase-strip" title="Engineering loop status">
      <span className="eng-phase-badge full">Engineering</span>
      <span className="eng-phase-name">{engineering.phase}</span>
      <div className="eng-evidence-chips">
        {EVIDENCE_KEYS.map(({ key, label }) => {
          const ok = !!engineering.evidence?.[key];
          return (
            <span key={key} className={`eng-chip ${ok ? 'ok' : 'missing'}`}>
              {label}
            </span>
          );
        })}
      </div>
    </div>
  );
};

export const ReviewCard: React.FC<{ review: NonNullable<EngineeringPhaseState['review']> }> = ({ review }) => {
  const verdict = (review.verdict || '').toLowerCase();
  const findings = review.findings || [];
  const issues = [
    ...(review.securityIssues || []),
    ...(findings || []),
  ];

  return (
    <div className={`review-card verdict-${verdict || 'unknown'}`}>
      <div className="review-header">
        <span className="review-title">Self-review</span>
        <span className="review-verdict">{verdict || 'n/a'}</span>
      </div>
      <div className="review-flags">
        {review.solvesProblem === false && <span className="eng-chip missing">problem unclear</span>}
        {review.unnecessaryComplexity && <span className="eng-chip missing">complexity</span>}
        {review.aiSlopRisk && <span className="eng-chip missing">ai-slop risk</span>}
      </div>
      {issues.length > 0 && (
        <ul className="review-findings">
          {issues.slice(0, 8).map((f, i) => (
            <li key={i}>{f}</li>
          ))}
        </ul>
      )}
    </div>
  );
};
