import React from 'react';
import {
  Activity,
  AlertOctagon,
  ArrowRight,
  CheckCircle2,
  CheckSquare,
  Clock,
  Radio,
  Server,
  ShieldAlert,
} from 'lucide-react';
import { Incident, OperationalEvent, Remediation } from '../types';

interface OverviewViewProps {
  incidents: Incident[];
  remediations: Remediation[];
  events: OperationalEvent[];
  onSelectIncident: (inc: Incident) => void;
  onNavigateTab: (tab: any) => void;
}

export const OverviewView: React.FC<OverviewViewProps> = ({
  incidents,
  remediations,
  events,
  onSelectIncident,
  onNavigateTab,
}) => {
  const openIncidents = incidents.filter((i) => i.status !== 'resolved');
  const criticalCount = openIncidents.filter((i) => i.severity === 'critical').length;
  const highCount = openIncidents.filter((i) => i.severity === 'high').length;
  const pendingRemediations = remediations.filter((r) => r.status === 'pending');

  return (
    <div>
      {/* Page Title */}
      <div style={{ marginBottom: '1.75rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
          Operations & Incident Intelligence Overview
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
          Real-time infrastructure health, automated event grouping, and active incident response pipeline.
        </p>
      </div>

      {/* KPI Cards Grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
        gap: '1.25rem',
        marginBottom: '2rem',
      }}>
        {/* Active Incidents */}
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)', fontWeight: 500 }}>Active Incidents</span>
            <div style={{ padding: '0.4rem', borderRadius: '8px', background: 'rgba(239, 68, 68, 0.15)' }}>
              <ShieldAlert size={18} color="#f87171" />
            </div>
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
            {openIncidents.length}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            <span style={{ color: '#f87171', fontWeight: 600 }}>{criticalCount} Critical</span> •{' '}
            <span style={{ color: '#fbbf24', fontWeight: 600 }}>{highCount} High</span>
          </div>
        </div>

        {/* Pending Remediations */}
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)', fontWeight: 500 }}>Remediations Awaiting Sign-off</span>
            <div style={{ padding: '0.4rem', borderRadius: '8px', background: 'rgba(245, 158, 11, 0.15)' }}>
              <CheckSquare size={18} color="#fbbf24" />
            </div>
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
            {pendingRemediations.length}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            Policy guardrails: Human approval required
          </div>
        </div>

        {/* Ingested Events */}
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)', fontWeight: 500 }}>Ingested Events (Stream)</span>
            <div style={{ padding: '0.4rem', borderRadius: '8px', background: 'rgba(6, 182, 212, 0.15)' }}>
              <Radio size={18} color="var(--accent-cyan)" />
            </div>
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
            {events.length}
          </div>
          <div style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            NATS JetStream persistent log
          </div>
        </div>

        {/* Resolved Incidents */}
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
            <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)', fontWeight: 500 }}>Resolved Incidents</span>
            <div style={{ padding: '0.4rem', borderRadius: '8px', background: 'rgba(16, 185, 129, 0.15)' }}>
              <CheckCircle2 size={18} color="#34d399" />
            </div>
          </div>
          <div style={{ fontSize: '2rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
            {incidents.filter((i) => i.status === 'resolved').length}
          </div>
          <div style={{ fontSize: '0.8rem', color: '#34d399' }}>
            100% verified resolution
          </div>
        </div>
      </div>

      {/* Main Grid: Active Incidents Table & Platform Pipeline */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '1.5rem' }}>
        {/* Active Incidents Spotlight */}
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem' }}>
            <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: '#fff' }}>Active Incidents Queue</h3>
            <button
              onClick={() => onNavigateTab('incidents')}
              className="btn btn-secondary"
              style={{ fontSize: '0.8rem', padding: '0.35rem 0.75rem' }}
            >
              <span>View All</span>
              <ArrowRight size={14} />
            </button>
          </div>

          {openIncidents.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '3rem 1rem', color: 'var(--text-muted)' }}>
              <CheckCircle2 size={36} color="var(--accent-green)" style={{ margin: '0 auto 0.75rem' }} />
              <p style={{ fontWeight: 500, color: '#fff' }}>No Active Incidents</p>
              <p style={{ fontSize: '0.85rem' }}>All systems operational. Incoming alerts will be correlated automatically.</p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              {openIncidents.slice(0, 5).map((inc) => (
                <div
                  key={inc.id}
                  onClick={() => onSelectIncident(inc)}
                  style={{
                    background: 'var(--bg-secondary)',
                    border: '1px solid var(--border-color)',
                    borderRadius: '0.5rem',
                    padding: '0.85rem 1.15rem',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    cursor: 'pointer',
                    transition: 'border-color 0.15s ease',
                  }}
                  onMouseEnter={(e) => (e.currentTarget.style.borderColor = 'var(--accent-cyan)')}
                  onMouseLeave={(e) => (e.currentTarget.style.borderColor = 'var(--border-color)')}
                >
                  <div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
                      <span className={`badge badge-${inc.severity}`}>{inc.severity}</span>
                      <span className={`badge badge-${inc.status}`}>{inc.status}</span>
                      <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>{inc.source}</span>
                    </div>
                    <div style={{ fontWeight: 600, fontSize: '0.95rem', color: '#fff' }}>{inc.title}</div>
                  </div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <Clock size={13} />
                    <span>{new Date(inc.createdAt).toLocaleTimeString()}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Intelligence Engine Status */}
        <div className="card">
          <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: '#fff', marginBottom: '1.25rem' }}>
            Incident Engine Stack
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', fontSize: '0.85rem' }}>
            <div style={{
              background: 'var(--bg-secondary)',
              padding: '0.85rem',
              borderRadius: '0.5rem',
              border: '1px solid var(--border-color)',
            }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.2rem' }}>
                1. Go Ingestion Layer
              </div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                Connectors for Prometheus, Alertmanager webhooks, and raw JSON payloads.
              </div>
            </div>

            <div style={{
              background: 'var(--bg-secondary)',
              padding: '0.85rem',
              borderRadius: '0.5rem',
              border: '1px solid var(--border-color)',
            }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.2rem' }}>
                2. NATS JetStream Broker
              </div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                Guaranteed event delivery across EVENTS, INCIDENTS, and ACTIONS streams.
              </div>
            </div>

            <div style={{
              background: 'var(--bg-secondary)',
              padding: '0.85rem',
              borderRadius: '0.5rem',
              border: '1px solid var(--border-color)',
            }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.2rem' }}>
                3. Deterministic RCA & AI Engine
              </div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                Algorithmic root-cause scoring with offline fallback; Ollama and OpenAI LLM ready.
              </div>
            </div>

            <div style={{
              background: 'var(--bg-secondary)',
              padding: '0.85rem',
              borderRadius: '0.5rem',
              border: '1px solid var(--border-color)',
            }}>
              <div style={{ fontWeight: 600, color: 'var(--accent-cyan)', marginBottom: '0.2rem' }}>
                4. Policy Guardrail & Rollback
              </div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                Mandatory human review before executing automated runbook remediations.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
