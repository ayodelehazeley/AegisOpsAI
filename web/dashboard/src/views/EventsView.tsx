import React, { useState } from 'react';
import {
  Code,
  Radio,
  Search,
} from 'lucide-react';
import { OperationalEvent } from '../types';

interface EventsViewProps {
  events: OperationalEvent[];
}

export const EventsView: React.FC<EventsViewProps> = ({ events }) => {
  const [selectedEvent, setSelectedEvent] = useState<OperationalEvent | null>(null);
  const [filterSource, setFilterSource] = useState('all');

  const sources = Array.from(new Set(events.map((e) => e.source)));
  const filtered = filterSource === 'all' ? events : events.filter((e) => e.source === filterSource);

  return (
    <div>
      {/* Title */}
      <div style={{ marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
          Real-Time Operational Event Stream
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
          Raw and normalized infrastructure events ingested via NATS JetStream from Prometheus, OpenTelemetry, and Webhooks.
        </p>
      </div>

      {/* Sources Filter */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.25rem' }}>
        <button
          onClick={() => setFilterSource('all')}
          className={`btn ${filterSource === 'all' ? 'btn-primary' : 'btn-secondary'}`}
          style={{ padding: '0.35rem 0.75rem', fontSize: '0.8rem' }}
        >
          All Sources ({events.length})
        </button>
        {sources.map((src) => (
          <button
            key={src}
            onClick={() => setFilterSource(src)}
            className={`btn ${filterSource === src ? 'btn-primary' : 'btn-secondary'}`}
            style={{ padding: '0.35rem 0.75rem', fontSize: '0.8rem' }}
          >
            {src} ({events.filter((e) => e.source === src).length})
          </button>
        ))}
      </div>

      {/* Events Table */}
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.875rem' }}>
          <thead>
            <tr style={{ background: 'var(--bg-secondary)', borderBottom: '1px solid var(--border-color)', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              <th style={{ padding: '1rem 1.25rem' }}>Severity</th>
              <th style={{ padding: '1rem 1.25rem' }}>Source</th>
              <th style={{ padding: '1rem 1.25rem' }}>Event Type</th>
              <th style={{ padding: '1rem 1.25rem' }}>Ingested Timestamp</th>
              <th style={{ padding: '1rem 1.25rem', textAlign: 'right' }}>Raw Payload</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ padding: '3rem 1.25rem', textAlign: 'center', color: 'var(--text-muted)' }}>
                  No events received yet. Use 'Simulate Incident' in the top bar to test ingestion.
                </td>
              </tr>
            ) : (
              filtered.map((ev) => (
                <tr
                  key={ev.id}
                  style={{ borderBottom: '1px solid var(--border-color)', cursor: 'pointer' }}
                  onClick={() => setSelectedEvent(ev)}
                  onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = 'var(--bg-card-hover)')}
                  onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
                >
                  <td style={{ padding: '1rem 1.25rem' }}>
                    <span className={`badge badge-${ev.severity}`}>{ev.severity}</span>
                  </td>
                  <td style={{ padding: '1rem 1.25rem', fontWeight: 600, color: '#fff' }}>
                    {ev.source}
                  </td>
                  <td style={{ padding: '1rem 1.25rem', color: 'var(--text-secondary)' }}>
                    {ev.type}
                  </td>
                  <td style={{ padding: '1rem 1.25rem', color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                    {new Date(ev.timestamp).toLocaleString()}
                  </td>
                  <td style={{ padding: '1rem 1.25rem', textAlign: 'right' }}>
                    <button className="btn btn-secondary" style={{ padding: '0.3rem 0.6rem', fontSize: '0.75rem' }}>
                      <Code size={13} />
                      <span>Inspect JSON</span>
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* JSON Payload Inspector Modal */}
      {selectedEvent && (
        <div className="modal-overlay" onClick={() => setSelectedEvent(null)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()} style={{ maxWidth: '650px', padding: '1.75rem' }}>
            <h3 style={{ fontSize: '1.1rem', fontWeight: 600, color: '#fff', marginBottom: '1rem' }}>
              Event Payload Inspector
            </h3>
            <pre style={{
              background: '#070a10',
              border: '1px solid #1e293b',
              borderRadius: '0.5rem',
              padding: '1rem',
              fontFamily: 'var(--font-mono)',
              fontSize: '0.825rem',
              color: '#38bdf8',
              overflowX: 'auto',
              maxHeight: '400px',
            }}>
              {JSON.stringify(selectedEvent.payload, null, 2)}
            </pre>
            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '1.25rem' }}>
              <button onClick={() => setSelectedEvent(null)} className="btn btn-secondary">
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
