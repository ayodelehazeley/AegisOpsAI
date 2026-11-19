import React from 'react';
import {
  Clock,
  FileCheck,
  Shield,
  User,
} from 'lucide-react';
import { AuditLog } from '../types';

interface AuditViewProps {
  logs: AuditLog[];
}

export const AuditView: React.FC<AuditViewProps> = ({ logs }) => {
  return (
    <div>
      {/* Title */}
      <div style={{ marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 700, color: '#fff', marginBottom: '0.25rem' }}>
          Immutable Security & Operational Audit Trail
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
          Cryptographically traceable records of engineer sign-offs, incident state changes, and automated remediation executions.
        </p>
      </div>

      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.875rem' }}>
          <thead>
            <tr style={{ background: 'var(--bg-secondary)', borderBottom: '1px solid var(--border-color)', color: 'var(--text-muted)', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              <th style={{ padding: '1rem 1.25rem' }}>Action</th>
              <th style={{ padding: '1rem 1.25rem' }}>Resource Type</th>
              <th style={{ padding: '1rem 1.25rem' }}>Resource Identifier</th>
              <th style={{ padding: '1rem 1.25rem' }}>Timestamp</th>
              <th style={{ padding: '1rem 1.25rem' }}>Audit Details</th>
            </tr>
          </thead>
          <tbody>
            {logs.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ padding: '3rem 1.25rem', textAlign: 'center', color: 'var(--text-muted)' }}>
                  No audit log entries recorded yet. Operational actions and sign-offs will appear here automatically.
                </td>
              </tr>
            ) : (
              logs.map((log) => (
                <tr key={log.id} style={{ borderBottom: '1px solid var(--border-color)' }}>
                  <td style={{ padding: '1rem 1.25rem' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '0.45rem', fontWeight: 600, color: 'var(--accent-cyan)' }}>
                      <FileCheck size={15} />
                      <span>{log.action}</span>
                    </div>
                  </td>
                  <td style={{ padding: '1rem 1.25rem', color: '#fff' }}>
                    {log.resourceType}
                  </td>
                  <td style={{ padding: '1rem 1.25rem', fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                    {log.resourceId || 'N/A'}
                  </td>
                  <td style={{ padding: '1rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>
                    {new Date(log.timestamp).toLocaleString()}
                  </td>
                  <td style={{ padding: '1rem 1.25rem', fontFamily: 'var(--font-mono)', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                    {JSON.stringify(log.details)}
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
