import React from 'react';
import {
  BrainCircuit,
  Check,
  Cpu,
  Database,
  Layers,
  Radio,
  Server,
  Shield,
} from 'lucide-react';

export const SettingsView: React.FC = () => {
  return (
    <div>
      {/* Title */}
      <div style={{ marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
          Platform Configuration & Architecture
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
          Open-source Go-first modular monolith architecture settings, AI provider integrations, and observability endpoints.
        </p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
        {/* AI Provider Architecture */}
        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem', marginBottom: '1rem', color: 'var(--accent-cyan)' }}>
            <BrainCircuit size={20} />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 600 }}>Provider-Independent AI</h3>
          </div>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '1.25rem', lineHeight: 1.5 }}>
            AegisOpsAI supports offline local models via <strong>Ollama</strong> as well as any <strong>OpenAI-compatible</strong> endpoint. Crucially, the platform's deterministic correlation, detection, and RCA algorithms function fully even when AI services are offline.
          </p>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', fontSize: '0.85rem' }}>
            <div style={{
              background: 'var(--bg-secondary)',
              border: '1px solid var(--border-color)',
              borderRadius: '0.5rem',
              padding: '0.85rem 1rem',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}>
              <div>
                <div style={{ fontWeight: 600, color: '#fff' }}>Local Inference (Ollama)</div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Default endpoint: http://localhost:11434</div>
              </div>
              <span className="badge badge-low">Supported</span>
            </div>

            <div style={{
              background: 'var(--bg-secondary)',
              border: '1px solid var(--border-color)',
              borderRadius: '0.5rem',
              padding: '0.85rem 1rem',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}>
              <div>
                <div style={{ fontWeight: 600, color: '#fff' }}>Cloud Inference (OpenAI / vLLM / LiteLLM)</div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Configured via AEGIS_AI_PROVIDER and AEGIS_AI_API_KEY</div>
              </div>
              <span className="badge badge-low">Supported</span>
            </div>

            <div style={{
              background: 'var(--bg-secondary)',
              border: '1px solid var(--border-color)',
              borderRadius: '0.5rem',
              padding: '0.85rem 1rem',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}>
              <div>
                <div style={{ fontWeight: 600, color: '#fff' }}>Deterministic Fallback Engine</div>
                <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Temporal cascade & resource exhaustion heuristics</div>
              </div>
              <span className="badge badge-medium">Active</span>
            </div>
          </div>
        </div>

        {/* Core Stack Architecture */}
        <div className="card">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem', marginBottom: '1rem', color: '#60a5fa' }}>
            <Layers size={20} />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 600 }}>Underlying Technology Stack</h3>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', fontSize: '0.85rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid var(--border-color)' }}>
              <span style={{ color: 'var(--text-muted)' }}>Backend Runtime:</span>
              <span style={{ fontWeight: 600, color: '#fff' }}>Go 1.23 Modular Monolith</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid var(--border-color)' }}>
              <span style={{ color: 'var(--text-muted)' }}>HTTP Router:</span>
              <span style={{ fontWeight: 600, color: '#fff' }}>Chi Router v5</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid var(--border-color)' }}>
              <span style={{ color: 'var(--text-muted)' }}>Primary Database:</span>
              <span style={{ fontWeight: 600, color: '#fff' }}>PostgreSQL 16 via pgx/v5</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid var(--border-color)' }}>
              <span style={{ color: 'var(--text-muted)' }}>Vector Search:</span>
              <span style={{ fontWeight: 600, color: '#fff' }}>pgvector (cosine similarity)</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid var(--border-color)' }}>
              <span style={{ color: 'var(--text-muted)' }}>Event Streaming Broker:</span>
              <span style={{ fontWeight: 600, color: '#fff' }}>NATS JetStream 2.10</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid var(--border-color)' }}>
              <span style={{ color: 'var(--text-muted)' }}>Observability:</span>
              <span style={{ fontWeight: 600, color: '#fff' }}>OpenTelemetry Go SDK</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0' }}>
              <span style={{ color: 'var(--text-muted)' }}>Frontend Dashboard:</span>
              <span style={{ fontWeight: 600, color: '#fff' }}>React 19 + TypeScript + Vite</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
