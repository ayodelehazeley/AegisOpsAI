import React, { useState } from 'react';
import {
  CheckCircle,
  CheckSquare,
  Play,
  RotateCcw,
  ShieldCheck,
  Terminal,
  UserCheck,
  Wrench,
} from 'lucide-react';
import { api } from '../api/client';
import { Remediation } from '../types';

interface RemediationsViewProps {
  remediations: Remediation[];
  onRemediationUpdated: () => void;
}

export const RemediationsView: React.FC<RemediationsViewProps> = ({
  remediations,
  onRemediationUpdated,
}) => {
  const [processingId, setProcessingId] = useState<string | null>(null);

  const handleApprove = async (id: string) => {
    setProcessingId(id);
    try {
      await api.approveRemediation(id);
      onRemediationUpdated();
    } catch (err: any) {
      alert(`Approval error: ${err.message}`);
    } finally {
      setProcessingId(null);
    }
  };

  const handleExecute = async (id: string) => {
    setProcessingId(id);
    try {
      await api.executeRemediation(id);
      onRemediationUpdated();
    } catch (err: any) {
      alert(`Execution error: ${err.message}`);
    } finally {
      setProcessingId(null);
    }
  };

  return (
    <div>
      {/* Title */}
      <div style={{ marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
          Remediation & Execution Hub
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
          Human-in-the-loop operational runbooks with deterministic verification and safety guardrails.
        </p>
      </div>

      {/* Safety Policy Notice */}
      <div style={{
        background: 'linear-gradient(90deg, rgba(59, 130, 246, 0.1) 0%, rgba(6, 182, 212, 0.1) 100%)',
        border: '1px solid rgba(59, 130, 246, 0.3)',
        borderRadius: '0.75rem',
        padding: '1.25rem 1.5rem',
        marginBottom: '1.75rem',
        display: 'flex',
        alignItems: 'center',
        gap: '1rem',
      }}>
        <div style={{ padding: '0.5rem', borderRadius: '50%', background: 'rgba(59, 130, 246, 0.2)' }}>
          <ShieldCheck size={24} color="#60a5fa" />
        </div>
        <div>
          <div style={{ fontWeight: 600, color: '#fff', marginBottom: '0.2rem' }}>
            Policy Guardrail Enforced: Mandatory Human Approval
          </div>
          <div style={{ fontSize: '0.825rem', color: 'var(--text-secondary)' }}>
            No autonomous change is executed against production infrastructure without an engineer reviewing the RCA evidence and providing explicit digital sign-off.
          </div>
        </div>
      </div>

      {/* Remediation Cards */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        {remediations.length === 0 ? (
          <div className="card" style={{ textAlign: 'center', padding: '3rem 1rem', color: 'var(--text-muted)' }}>
            <Wrench size={36} color="var(--accent-cyan)" style={{ margin: '0 auto 0.75rem' }} />
            <p style={{ fontWeight: 500, color: '#fff' }}>No Remediations Currently Queued</p>
            <p style={{ fontSize: '0.85rem' }}>
              When an incident is investigated, click 'Propose Controlled Remediation' in the incident details to queue a runbook here.
            </p>
          </div>
        ) : (
          remediations.map((rem) => {
            const isProcessing = processingId === rem.id;
            return (
              <div key={rem.id} className="card">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '1rem' }}>
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem', marginBottom: '0.35rem' }}>
                      <span className={`badge badge-${rem.status === 'completed' ? 'low' : rem.status === 'approved' ? 'medium' : rem.status === 'pending' ? 'high' : 'critical'}`}>
                        {rem.status}
                      </span>
                      <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        Action: <strong style={{ color: '#fff' }}>{rem.type}</strong>
                      </span>
                      <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        Queued: {new Date(rem.createdAt).toLocaleTimeString()}
                      </span>
                    </div>
                    <div style={{ fontWeight: 600, fontSize: '1rem', color: '#fff' }}>
                      {rem.description}
                    </div>
                  </div>

                  {/* Actions */}
                  <div style={{ display: 'flex', gap: '0.75rem' }}>
                    {rem.status === 'pending' && (
                      <button
                        disabled={isProcessing}
                        onClick={() => handleApprove(rem.id)}
                        className="btn btn-primary"
                        style={{ padding: '0.45rem 0.9rem', fontSize: '0.8rem' }}
                      >
                        <UserCheck size={15} />
                        <span>{isProcessing ? 'Approving...' : 'Sign Off & Approve'}</span>
                      </button>
                    )}

                    {rem.status === 'approved' && (
                      <button
                        disabled={isProcessing}
                        onClick={() => handleExecute(rem.id)}
                        className="btn btn-success"
                        style={{ padding: '0.45rem 0.9rem', fontSize: '0.8rem' }}
                      >
                        <Play size={15} />
                        <span>{isProcessing ? 'Executing...' : 'Execute Runbook'}</span>
                      </button>
                    )}

                    {rem.status === 'completed' && (
                      <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '0.35rem',
                        fontSize: '0.8rem',
                        color: '#34d399',
                        padding: '0.4rem 0.75rem',
                        borderRadius: '0.5rem',
                        background: 'rgba(16, 185, 129, 0.1)',
                      }}>
                        <CheckCircle size={15} />
                        <span>Verified Resolved</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Execution Output Console */}
                {rem.result && (
                  <div style={{
                    background: '#070a10',
                    border: '1px solid #1e293b',
                    borderRadius: '0.5rem',
                    padding: '0.85rem 1rem',
                    fontFamily: 'var(--font-mono)',
                    fontSize: '0.8rem',
                    color: '#94a3b8',
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', color: 'var(--accent-cyan)', marginBottom: '0.35rem', fontSize: '0.75rem' }}>
                      <Terminal size={13} />
                      <span>Execution Output Log:</span>
                    </div>
                    <div style={{ color: '#f1f5f9' }}>{rem.result}</div>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
