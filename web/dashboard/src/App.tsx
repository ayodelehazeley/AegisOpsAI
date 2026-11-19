import React, { useEffect, useState } from 'react';
import { api } from './api/client';
import { IncidentDetailModal } from './components/IncidentDetailModal';
import { Navbar } from './components/Navbar';
import { NavTab, Sidebar } from './components/Sidebar';
import {
  AuditLog,
  Incident,
  OperationalEvent,
  Remediation,
  SystemHealth,
} from './types';
import { AuditView } from './views/AuditView';
import { EventsView } from './views/EventsView';
import { IncidentsView } from './views/IncidentsView';
import { OverviewView } from './views/OverviewView';
import { RemediationsView } from './views/RemediationsView';
import { SettingsView } from './views/SettingsView';

export function App() {
  const [activeTab, setActiveTab] = useState<NavTab>('overview');
  const [health, setHealth] = useState<SystemHealth | null>(null);
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [remediations, setRemediations] = useState<Remediation[]>([]);
  const [events, setEvents] = useState<OperationalEvent[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    try {
      const [h, incs, evts] = await Promise.all([
        api.getHealth().catch(() => null),
        api.getIncidents().catch(() => []),
        api.getEvents().catch(() => []),
      ]);
      if (h) setHealth(h);
      if (incs) setIncidents(incs);
      if (evts) setEvents(evts);

      // Audit logs (optional admin query)
      try {
        const logs = await api.getAuditLogs();
        setAuditLogs(logs);
      } catch {
        // may require admin auth token
      }
    } catch (err) {
      console.error('Failed to poll AegisOpsAI telemetry:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 6000);
    return () => clearInterval(interval);
  }, []);

  const openIncidentCount = incidents.filter((i) => i.status !== 'resolved').length;
  const pendingRemediationsCount = remediations.filter((r) => r.status === 'pending').length;

  return (
    <div className="app-container">
      {/* Sidebar Navigation */}
      <Sidebar
        activeTab={activeTab}
        onTabChange={setActiveTab}
        incidentCount={openIncidentCount}
        pendingRemediationsCount={pendingRemediationsCount}
      />

      {/* Main Content Area */}
      <div className="main-content">
        <Navbar
          health={health}
          onRefresh={fetchData}
          onEventSimulated={() => {
            fetchData();
            setTimeout(fetchData, 1000);
          }}
        />

        <main className="page-body">
          {activeTab === 'overview' && (
            <OverviewView
              incidents={incidents}
              remediations={remediations}
              events={events}
              onSelectIncident={(inc) => setSelectedIncident(inc)}
              onNavigateTab={setActiveTab}
            />
          )}

          {activeTab === 'incidents' && (
            <IncidentsView
              incidents={incidents}
              onSelectIncident={(inc) => setSelectedIncident(inc)}
            />
          )}

          {activeTab === 'remediations' && (
            <RemediationsView
              remediations={remediations}
              onRemediationUpdated={fetchData}
            />
          )}

          {activeTab === 'events' && (
            <EventsView events={events} />
          )}

          {activeTab === 'audit' && (
            <AuditView logs={auditLogs} />
          )}

          {activeTab === 'settings' && (
            <SettingsView />
          )}
        </main>
      </div>

      {/* Incident Detail / Investigation Modal */}
      {selectedIncident && (
        <IncidentDetailModal
          incident={selectedIncident}
          events={events.filter((e) => e.incidentId === selectedIncident.id)}
          onClose={() => setSelectedIncident(null)}
          onStatusUpdated={() => {
            fetchData();
            setSelectedIncident(null);
          }}
        />
      )}
    </div>
  );
}

export default App;
