import React, { useState } from 'react';
import {
  ChevronRight,
  Filter,
  Search,
  ShieldAlert,
} from 'lucide-react';
import { Incident, IncidentStatus, Severity } from '../types';

interface IncidentsViewProps {
  incidents: Incident[];
  onSelectIncident: (inc: Incident) => void;
}

export const IncidentsView: React.FC<IncidentsViewProps> = ({ incidents, onSelectIncident }) => {
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [severityFilter, setSeverityFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');

  const filtered = incidents.filter((inc) => {
    if (statusFilter !== 'all' && inc.status !== statusFilter) return false;
    if (severityFilter !== 'all' && inc.severity !== severityFilter) return false;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      return inc.title.toLowerCase().includes(q) || inc.source.toLowerCase().includes(q);
    }
    return true;
  });

  return (
    <div>
      {/* Title */}
      <div style={{ marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
          Incident Management Console
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
          Correlated operational incidents grouped by service topology, temporal proximity, and causal evidence.
        </p>
      </div>

      {/* Filter and Search Bar */}
      <div style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: '1rem',
        alignItems: 'center',
        justifyContent: 'space-between',
        background: 'var(--bg-card)',
        padding: '1rem 1.25rem',
        borderRadius: '0.75rem',
        border: '1px solid var(--border-color)',
        marginBottom: '1.5rem',
      }}>
        {/* Search */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          background: 'var(--bg-secondary)',
          border: '1px solid var(--border-color)',
          borderRadius: '0.5rem',
          padding: '0.45rem 0.75rem',
          minWidth: '280px',
        }}>
          <Search size={16} color="var(--text-muted)" />
          <input
            type="text"
            placeholder="Search by title, service, source..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{
              background: 'none',
              border: 'none',
              outline: 'none',
              color: '#fff',
              fontSize: '0.85rem',
              width: '100%',
            }}
          />
        </div>

        {/* Filters */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
            <span style={{ color: 'var(--text-muted)' }}>Status:</span>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              style={{
                background: 'var(--bg-secondary)',
                border: '1px solid var(--border-color)',
                color: '#fff',
                padding: '0.35rem 0.6rem',
                borderRadius: '0.4rem',
                fontSize: '0.8rem',
                outline: 'none',
              }}
            >
              <option value="all">All Statuses</option>
              <option value="open">Open</option>
              <option value="investigating">Investigating</option>
              <option value="identified">Identified</option>
              <option value="monitoring">Monitoring</option>
              <option value="resolved">Resolved</option>
            </select>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.85rem' }}>
            <span style={{ color: 'var(--text-muted)' }}>Severity:</span>
            <select
              value={severityFilter}
              onChange={(e) => setSeverityFilter(e.target.value)}
              style={{
                background: 'var(--bg-secondary)',
                border: '1px solid var(--border-color)',
                color: '#fff',
                padding: '0.35rem 0.6rem',
                borderRadius: '0.4rem',
                fontSize: '0.8rem',
                outline: 'none',
              }}
            >
              <option value="all">All Severities</option>
              <option value="critical">Critical</option>
              <option value="high">High</option>
              <option value="medium">Medium</option>
              <option value="low">Low</option>
            </select>
          </div>
        </div>
      </div>

      {/* Incidents List Card */}
      <div className="card" style={{ padding: '0', overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.875rem' }}>
          <thead>
            <tr style={{ background: 'var(--bg-secondary)', borderBottom: '1px solid var(--border-color)', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              <th style={{ padding: '1rem 1.25rem' }}>Severity</th>
              <th style={{ padding: '1rem 1.25rem' }}>Title & Description</th>
              <th style={{ padding: '1rem 1.25rem' }}>Status</th>
              <th style={{ padding: '1rem 1.25rem' }}>Source</th>
              <th style={{ padding: '1rem 1.25rem' }}>Created</th>
              <th style={{ padding: '1rem 1.25rem', textAlign: 'right' }}>Action</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ padding: '3rem 1.25rem', textAlign: 'center', color: 'var(--text-muted)' }}>
                  No matching incidents found.
                </td>
              </tr>
            ) : (
              filtered.map((inc) => (
                <tr
                  key={inc.id}
                  onClick={() => onSelectIncident(inc)}
                  style={{
                    borderBottom: '1px solid var(--border-color)',
                    cursor: 'pointer',
                    transition: 'background-color 0.15s ease',
                  }}
                  onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = 'var(--bg-card-hover)')}
                  onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
                >
                  <td style={{ padding: '1rem 1.25rem' }}>
                    <span className={`badge badge-${inc.severity}`}>{inc.severity}</span>
                  </td>
                  <td style={{ padding: '1rem 1.25rem' }}>
                    <div style={{ fontWeight: 600, color: '#fff', marginBottom: '0.2rem' }}>{inc.title}</div>
                    <div style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', maxWidth: '420px', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                      {inc.description || 'No description provided'}
                    </div>
                  </td>
                  <td style={{ padding: '1rem 1.25rem' }}>
                    <span className={`badge badge-${inc.status}`}>{inc.status}</span>
                  </td>
                  <td style={{ padding: '1rem 1.25rem', color: 'var(--text-secondary)' }}>
                    {inc.source}
                  </td>
                  <td style={{ padding: '1rem 1.25rem', color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                    {new Date(inc.createdAt).toLocaleString()}
                  </td>
                  <td style={{ padding: '1rem 1.25rem', textAlign: 'right' }}>
                    <button
                      className="btn btn-secondary"
                      style={{ padding: '0.35rem 0.75rem', fontSize: '0.75rem' }}
                      onClick={(e) => {
                        e.stopPropagation();
                        onSelectIncident(inc);
                      }}
                    >
                      <span>Investigate</span>
                      <ChevronRight size={14} />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
