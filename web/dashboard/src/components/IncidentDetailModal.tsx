import React, { useState } from 'react';
import {
  AlertCircle,
  AlertTriangle,
  BrainCircuit,
  CheckCircle,
  Clock,
  Layers,
  Sparkles,
  Wrench,
  X,
} from 'lucide-react';
import { api } from '../api/client';
import { Incident, IncidentStatus, OperationalEvent } from '../types';

interface IncidentDetailModalProps {
  incident: Incident;
  events: OperationalEvent[];
  onClose: () => void;
  onStatusUpdated: () => void;
}

export const IncidentDetailModal: React.FC<IncidentDetailModalProps> = ({
  incident,
  events,
  onClose,
  onStatusUpdated,
}) => {
  const [updating, setUpdating] = useState(false);
  const [proposing, setProposing] = useState(false);
  const [proposedMsg, setProposedMsg] = useState('');

  const rca = api.computeRCA(incident, events);
  const aiSummary = api.generateSummary(incident, rca);

  const handleStatusChange = async (newStatus: IncidentStatus) => {
    setUpdating(true);
    try {
      await api.updateIncidentStatus(incident.id, newStatus);
      onStatusUpdated();
    } catch (err: any) {
      alert(`Failed to update status: ${err.message}`);
    } finally {
      setUpdating(false);
    }
  };

  const handleProposeRemediation = async () => {
    setProposing(true);
    try {
      const actionType = incident.title.toLowerCase().includes('disk') ? 'ClearDiskSpace' : 'RestartService';
      await api.proposeRemediation(
        incident.id,
        actionType,
        `Remediation proposed for ${incident.title}: Execute ${actionType} with outcome verification.`,
      );
      setProposedMsg(`Remediation action '${actionType}' proposed! Switched to Remediation Hub for human approval.`);
    } catch (err: any) {
      alert(`Failed to propose remediation: ${err.message}`);
    } finally {
      setProposing(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ padding: '2rem' }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem', marginBottom: '0.5rem' }}>
              <span className={`badge badge-${incident.severity}`}>{incident.severity}</span>
              <span className={`badge badge-${incident.status}`}>{incident.status}</span>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Source: {incident.source}</span>
            </div>
            <h2 style={{ fontSize: '1.4rem', fontWeight: 700, color: '#fff' }}>{incident.title}</h2>
          </div>
          <button onClick={onClose} className="btn btn-secondary" style={{ padding: '0.4rem', borderRadius: '50%' }}>
            <X size={18} />
          </button>
        </div>

        {/* Incident Description */}
        {incident.description && (
          <div style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border-color)',
            borderRadius: '0.5rem',
            padding: '1rem',
            marginBottom: '1.5rem',
            fontSize: '0.9rem',
            color: 'var(--text-secondary)',
          }}>
            {incident.description}
          </div>
        )}

        {/* AI & RCA Grid */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.25rem', marginBottom: '1.75rem' }}>
          {/* AI Incident Intelligence Card */}
          <div style={{
            background: 'linear-gradient(180deg, rgba(6, 182, 212, 0.08) 0%, rgba(22, 31, 48, 0.6) 100%)',
            border: '1px solid rgba(6, 182, 212, 0.25)',
            borderRadius: '0.75rem',
            padding: '1.25rem',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--accent-cyan)' }}>
                <BrainCircuit size={18} />
                <h4 style={{ fontWeight: 600, fontSize: '0.95rem' }}>AI Incident Intelligence</h4>
              </div>
              <span style={{
                fontSize: '0.65rem',
                padding: '0.15rem 0.45rem',
                background: 'rgba(6, 182, 212, 0.15)',
                color: 'var(--accent-cyan)',
                borderRadius: '4px',
                fontWeight: 600,
                display: 'flex',
                alignItems: 'center',
                gap: '0.25rem',
              }}>
                <Sparkles size={11} />
                LLM Inference
              </span>
            </div>

            <p style={{ fontSize: '0.85rem', color: '#e2e8f0', marginBottom: '0.85rem', lineHeight: 1.5 }}>
              {aiSummary.summary}
            </p>

            <div style={{ marginBottom: '0.85rem' }}>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: '0.35rem', textTransform: 'uppercase' }}>
                Key Evidence Findings
              </div>
              <ul style={{ paddingLeft: '1.2rem', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                {aiSummary.keyFindings.map((finding, idx) => (
                  <li key={idx} style={{ marginBottom: '0.2rem' }}>{finding}</li>
                ))}
              </ul>
            </div>

            <div>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-muted)', marginBottom: '0.35rem', textTransform: 'uppercase' }}>
                Recommended Runbooks
              </div>
              <ul style={{ paddingLeft: '1.2rem', fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                {aiSummary.recommendedActions.map((action, idx) => (
                  <li key={idx} style={{ marginBottom: '0.2rem' }}>{action}</li>
                ))}
              </ul>
            </div>
          </div>

          {/* Deterministic RCA Engine Card */}
          <div style={{
            background: 'var(--bg-card)',
            border: '1px solid var(--border-color)',
            borderRadius: '0.75rem',
            padding: '1.25rem',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#fbbf24' }}>
                <Layers size={18} />
                <h4 style={{ fontWeight: 600, fontSize: '0.95rem' }}>Deterministic RCA Engine</h4>
              </div>
              <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                Confidence: <strong style={{ color: '#fff' }}>{(rca.overallScore * 100).toFixed(0)}%</strong>
              </span>
            </div>

            <p style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.85rem' }}>
              Rule-based heuristic analysis: temporal cascades and resource saturation signatures.
            </p>

            {rca.probableCauses.map((cause, idx) => (
              <div key={idx} style={{
                background: 'var(--bg-secondary)',
                borderRadius: '0.5rem',
                padding: '0.75rem',
                marginBottom: '0.75rem',
                border: '1px solid var(--border-color)',
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.25rem' }}>
                  <span style={{ fontWeight: 600, fontSize: '0.85rem', color: '#fff' }}>{cause.component}</span>
                  <span style={{ fontSize: '0.75rem', color: '#fbbf24', fontWeight: 600 }}>{(cause.likelihood * 100).toFixed(0)}% Likelihood</span>
                </div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.4rem' }}>
                  Failure Mode: {cause.failureType}
                </div>
                <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{cause.description}</p>
              </div>
            ))}
          </div>
        </div>

        {/* Proposed Remediation Alert Message */}
        {proposedMsg && (
          <div style={{
            background: 'rgba(16, 185, 129, 0.15)',
            border: '1px solid rgba(16, 185, 129, 0.4)',
            color: '#34d399',
            padding: '0.75rem 1rem',
            borderRadius: '0.5rem',
            fontSize: '0.85rem',
            marginBottom: '1.25rem',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}>
            <CheckCircle size={16} />
            <span>{proposedMsg}</span>
          </div>
        )}

        {/* Operational Actions Footer */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          paddingTop: '1.25rem',
          borderTop: '1px solid var(--border-color)',
        }}>
          {/* Remediation Proposal Button */}
          <button
            disabled={proposing}
            onClick={handleProposeRemediation}
            className="btn btn-primary"
            style={{ gap: '0.5rem' }}
          >
            <Wrench size={16} />
            <span>{proposing ? 'Proposing...' : 'Propose Controlled Remediation'}</span>
          </button>

          {/* Status Controls */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Transition Status:</span>
            {(['open', 'investigating', 'identified', 'monitoring', 'resolved'] as IncidentStatus[]).map((st) => (
              <button
                key={st}
                disabled={updating || incident.status === st}
                onClick={() => handleStatusChange(st)}
                className={`btn ${incident.status === st ? 'btn-primary' : 'btn-secondary'}`}
                style={{
                  fontSize: '0.75rem',
                  padding: '0.35rem 0.65rem',
                  opacity: incident.status === st ? 0.6 : 1,
                  textTransform: 'capitalize',
                }}
              >
                {st}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
