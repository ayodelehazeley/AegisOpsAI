import React from 'react';
import {
  Activity,
  AlertTriangle,
  CheckSquare,
  Cpu,
  FileText,
  Radio,
  Settings,
  ShieldAlert,
} from 'lucide-react';

export type NavTab = 'overview' | 'incidents' | 'remediations' | 'events' | 'audit' | 'settings';

interface SidebarProps {
  activeTab: NavTab;
  onTabChange: (tab: NavTab) => void;
  incidentCount: number;
  pendingRemediationsCount: number;
}

export const Sidebar: React.FC<SidebarProps> = ({
  activeTab,
  onTabChange,
  incidentCount,
  pendingRemediationsCount,
}) => {
  const navItems = [
    { id: 'overview', label: 'Overview', icon: Activity },
    { id: 'incidents', label: 'Incidents', icon: ShieldAlert, badge: incidentCount },
    { id: 'remediations', label: 'Remediations', icon: CheckSquare, badge: pendingRemediationsCount },
    { id: 'events', label: 'Event Stream', icon: Radio },
    { id: 'audit', label: 'Audit Trail', icon: FileText },
    { id: 'settings', label: 'AI & Architecture', icon: Settings },
  ];

  return (
    <aside style={{
      width: '260px',
      background: 'var(--bg-secondary)',
      borderRight: '1px solid var(--border-color)',
      display: 'flex',
      flexDirection: 'column',
      padding: '1.5rem 1rem',
      flexShrink: 0,
    }}>
      {/* Brand Header */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '0.75rem',
        padding: '0 0.5rem 1.75rem',
        borderBottom: '1px solid var(--border-color)',
        marginBottom: '1.5rem',
      }}>
        <div style={{
          width: '38px',
          height: '38px',
          borderRadius: '10px',
          background: 'linear-gradient(135deg, #06b6d4 0%, #3b82f6 100%)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: '0 4px 12px rgba(6, 182, 212, 0.35)',
        }}>
          <AlertTriangle size={22} color="#ffffff" />
        </div>
        <div>
          <h1 style={{ fontSize: '1.15rem', fontWeight: 700, letterSpacing: '-0.02em', color: '#fff' }}>
            AegisOps<span style={{ color: 'var(--accent-cyan)' }}>AI</span>
          </h1>
          <p style={{ fontSize: '0.7rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>
            Incident Intelligence
          </p>
        </div>
      </div>

      {/* Navigation Items */}
      <nav style={{ display: 'flex', flexDirection: 'column', gap: '0.35rem', flex: 1 }}>
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = activeTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => onTabChange(item.id as NavTab)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.85rem',
                padding: '0.75rem 1rem',
                borderRadius: '0.6rem',
                fontSize: '0.9rem',
                fontWeight: isActive ? 600 : 500,
                color: isActive ? '#fff' : 'var(--text-secondary)',
                backgroundColor: isActive ? 'var(--bg-card)' : 'transparent',
                border: isActive ? '1px solid var(--border-highlight)' : '1px solid transparent',
                transition: 'all 0.15s ease',
                textAlign: 'left',
                width: '100%',
              }}
            >
              <Icon size={18} color={isActive ? 'var(--accent-cyan)' : 'var(--text-muted)'} />
              <span style={{ flex: 1 }}>{item.label}</span>
              {item.badge !== undefined && item.badge > 0 && (
                <span style={{
                  padding: '0.15rem 0.5rem',
                  borderRadius: '9999px',
                  fontSize: '0.7rem',
                  fontWeight: 700,
                  background: item.id === 'incidents' ? 'rgba(239, 68, 68, 0.2)' : 'rgba(245, 158, 11, 0.2)',
                  color: item.id === 'incidents' ? '#f87171' : '#fbbf24',
                }}>
                  {item.badge}
                </span>
              )}
            </button>
          );
        })}
      </nav>

      {/* Footer Info */}
      <div style={{
        paddingTop: '1rem',
        borderTop: '1px solid var(--border-color)',
        fontSize: '0.75rem',
        color: 'var(--text-muted)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', marginBottom: '0.25rem' }}>
          <Cpu size={14} color="var(--accent-cyan)" />
          <span>Core: Go Monolith v0.1</span>
        </div>
        <p>Apache 2.0 Open-Source</p>
      </div>
    </aside>
  );
};
