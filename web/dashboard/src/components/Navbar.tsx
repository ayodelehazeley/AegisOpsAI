import React, { useState } from 'react';
import {
  Bell,
  CheckCircle2,
  Database,
  Flame,
  Radio,
  RefreshCw,
  Server,
  Zap,
} from 'lucide-react';
import { api } from '../api/client';
import { SystemHealth } from '../types';

interface NavbarProps {
  health: SystemHealth | null;
  onRefresh: () => void;
  onEventSimulated: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ health, onRefresh, onEventSimulated }) => {
  const [simulating, setSimulating] = useState(false);
  const [showSimModal, setShowSimModal] = useState(false);

  const simulateIncident = async (type: 'disk' | 'memory' | 'latency') => {
    setSimulating(true);
    try {
      if (type === 'disk') {
        await api.sendWebhookEvent({
          source: 'prometheus-alertmanager',
          title: 'DiskVolumeSaturationThresholdExceeded',
          severity: 'critical',
          description: 'Root disk filesystem storage utilization reached 96.8% on host node-prod-04. Write IO blocked.',
        });
      } else if (type === 'memory') {
        await api.sendWebhookEvent({
          source: 'prometheus-alertmanager',
          title: 'KernelOOMKillImminent',
          severity: 'high',
          description: 'Cgroup memory threshold breached on api-gateway-pod. Heap memory consumption growing at 12MB/sec.',
        });
      } else {
        await api.sendWebhookEvent({
          source: 'opentelemetry-collector',
          title: 'P99ServiceLatencyBreach',
          severity: 'high',
          description: 'Ingress HTTP gateway P99 latency exceeded SLA of 2000ms (current: 4120ms).',
        });
      }
      setShowSimModal(false);
      onEventSimulated();
    } catch (err: any) {
      alert(`Simulation failed: ${err.message}`);
    } finally {
      setSimulating(false);
    }
  };

  const isHealthy = health?.status === 'healthy';

  return (
    <header style={{
      height: '64px',
      background: 'var(--bg-secondary)',
      borderBottom: '1px solid var(--border-color)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '0 2rem',
      position: 'sticky',
      top: 0,
      zIndex: 100,
    }}>
      {/* Platform & Cluster Context */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.875rem' }}>
          <Server size={16} color="var(--accent-cyan)" />
          <span style={{ color: 'var(--text-muted)' }}>Cluster:</span>
          <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>default-local (Docker)</span>
        </div>

        {/* System Health Indicators */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.4rem',
            padding: '0.2rem 0.6rem',
            background: isHealthy ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)',
            border: `1px solid ${isHealthy ? 'rgba(16, 185, 129, 0.3)' : 'rgba(239, 68, 68, 0.3)'}`,
            borderRadius: '9999px',
            fontSize: '0.75rem',
            color: isHealthy ? '#34d399' : '#f87171',
          }}>
            <Database size={12} />
            <span>PostgreSQL: {health?.components?.find(c => c.name === 'postgresql')?.status || 'ready'}</span>
          </div>

          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.4rem',
            padding: '0.2rem 0.6rem',
            background: 'rgba(6, 182, 212, 0.1)',
            border: '1px solid rgba(6, 182, 212, 0.3)',
            borderRadius: '9999px',
            fontSize: '0.75rem',
            color: 'var(--accent-cyan)',
          }}>
            <Radio size={12} />
            <span>NATS JetStream: {health?.components?.find(c => c.name === 'nats')?.status || 'ready'}</span>
          </div>
        </div>
      </div>

      {/* Action Controls */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
        {/* Simulate Incident Button */}
        <button
          onClick={() => setShowSimModal(true)}
          className="btn btn-primary"
          style={{ padding: '0.4rem 0.85rem', fontSize: '0.8rem' }}
        >
          <Flame size={15} />
          <span>Simulate Incident</span>
        </button>

        {/* Refresh Button */}
        <button
          onClick={onRefresh}
          className="btn btn-secondary"
          style={{ padding: '0.45rem', borderRadius: '50%' }}
          title="Refresh operational telemetry"
        >
          <RefreshCw size={16} />
        </button>
      </div>

      {/* Simulation Modal */}
      {showSimModal && (
        <div className="modal-overlay" onClick={() => setShowSimModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '520px', padding: '1.75rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1rem' }}>
              <Zap size={22} color="var(--accent-cyan)" />
              <h3 style={{ fontSize: '1.15rem', fontWeight: 600 }}>Inject Operational Incident</h3>
            </div>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem', marginBottom: '1.25rem' }}>
              Select an infrastructure event to dispatch to the Go Ingestion Pipeline. The correlation and RCA engine will process the event in real time.
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              <button
                disabled={simulating}
                onClick={() => simulateIncident('disk')}
                className="btn btn-secondary"
                style={{ justifyContent: 'flex-start', padding: '0.85rem 1rem', textAlign: 'left' }}
              >
                <div>
                  <div style={{ fontWeight: 600, color: '#f87171' }}>Critical: Disk Saturation (96.8%)</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Source: Prometheus • triggers ClearDiskSpace runbook suggestion</div>
                </div>
              </button>

              <button
                disabled={simulating}
                onClick={() => simulateIncident('memory')}
                className="btn btn-secondary"
                style={{ justifyContent: 'flex-start', padding: '0.85rem 1rem', textAlign: 'left' }}
              >
                <div>
                  <div style={{ fontWeight: 600, color: '#fbbf24' }}>High: Kernel OOM Kill Warning</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Source: Node Exporter • triggers RestartService runbook suggestion</div>
                </div>
              </button>

              <button
                disabled={simulating}
                onClick={() => simulateIncident('latency')}
                className="btn btn-secondary"
                style={{ justifyContent: 'flex-start', padding: '0.85rem 1rem', textAlign: 'left' }}
              >
                <div>
                  <div style={{ fontWeight: 600, color: '#60a5fa' }}>High: Ingress P99 Latency Breach</div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Source: OpenTelemetry • evaluates cascading gateway downstream delays</div>
                </div>
              </button>
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '1.5rem' }}>
              <button onClick={() => setShowSimModal(false)} className="btn btn-secondary">
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </header>
  );
};
