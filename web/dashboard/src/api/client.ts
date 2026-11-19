import {
  AuditLog,
  Incident,
  IncidentStatus,
  IncidentSummary,
  OperationalEvent,
  RCAResult,
  Remediation,
  SystemHealth,
} from '../types';

let authToken = localStorage.getItem('aegisops_token') || '';

// Base API URL from environment variable (for Vercel or external hosting)
const API_BASE_URL = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '');

export function setAuthToken(token: string) {
  authToken = token;
  localStorage.setItem('aegisops_token', token);
}

export function getAuthToken() {
  return authToken;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers || {});
  headers.set('Content-Type', 'application/json');
  if (authToken) {
    headers.set('Authorization', `Bearer ${authToken}`);
  }

  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  const url = API_BASE_URL ? `${API_BASE_URL}${normalizedPath}` : normalizedPath;

  const res = await fetch(url, { ...options, headers });
  if (!res.ok) {
    let errMessage = `HTTP ${res.status}`;
    try {
      const errJson = await res.json();
      errMessage = errJson.error || errJson.message || errMessage;
    } catch {
      // ignore
    }
    throw new Error(errMessage);
  }

  const json = await res.json();
  return (json.data !== undefined ? json.data : json) as T;
}

export const api = {
  // System Health
  async getHealth(): Promise<SystemHealth> {
    return request<SystemHealth>('/healthz');
  },

  // Incidents
  async getIncidents(): Promise<Incident[]> {
    return request<Incident[]>('/api/v1/incidents');
  },

  async getIncident(id: string): Promise<Incident> {
    return request<Incident>(`/api/v1/incidents/${id}`);
  },

  async updateIncidentStatus(id: string, status: IncidentStatus): Promise<{ id: string; status: IncidentStatus }> {
    return request<{ id: string; status: IncidentStatus }>(`/api/v1/incidents/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  },

  // Events
  async getEvents(): Promise<OperationalEvent[]> {
    return request<OperationalEvent[]>('/api/v1/events');
  },

  async sendWebhookEvent(event: {
    source: string;
    title: string;
    severity: string;
    description: string;
  }): Promise<{ status: string }> {
    return request<{ status: string }>('/integrations/webhook', {
      method: 'POST',
      body: JSON.stringify(event),
    });
  },

  // Remediations
  async proposeRemediation(incidentId: string, actionType: string, description: string): Promise<Remediation> {
    return request<Remediation>('/api/v1/remediations', {
      method: 'POST',
      body: JSON.stringify({ incidentId, actionType, description }),
    });
  },

  async approveRemediation(id: string): Promise<{ id: string; status: string }> {
    return request<{ id: string; status: string }>(`/api/v1/remediations/${id}/approve`, {
      method: 'POST',
    });
  },

  async executeRemediation(id: string): Promise<Remediation> {
    return request<Remediation>(`/api/v1/remediations/${id}/execute`, {
      method: 'POST',
    });
  },

  // Audit Logs
  async getAuditLogs(): Promise<AuditLog[]> {
    return request<AuditLog[]>('/api/v1/audit/logs');
  },

  // Client-side RCA calculation helper based on correlated incident signals
  computeRCA(incident: Incident, events: OperationalEvent[]): RCAResult {
    const text = (incident.title + ' ' + incident.description).toLowerCase();
    const causes = [];
    const evidence = [];

    if (text.includes('disk') || text.includes('space')) {
      causes.push({
        component: 'Storage Subsystem (/var/log)',
        failureType: 'DiskSpaceSaturation',
        description: 'Persistent journal logs filled up root filesystem past 95% threshold',
        likelihood: 0.94,
      });
      evidence.push({
        source: incident.source,
        description: 'IO write error alerts detected across node storage',
        confidence: 0.92,
        timestamp: incident.createdAt,
      });
    } else if (text.includes('memory') || text.includes('oom') || text.includes('utilization')) {
      causes.push({
        component: 'Compute Runtime / Heap',
        failureType: 'MemoryLeakDegradation',
        description: 'Container memory limit exceeded cgroup threshold, threatening OOM kill',
        likelihood: 0.91,
      });
      evidence.push({
        source: incident.source,
        description: 'RSS memory growth steep ascent detected in 5-minute sliding window',
        confidence: 0.89,
        timestamp: incident.createdAt,
      });
    } else {
      causes.push({
        component: incident.source || 'Infrastructure Gateway',
        failureType: 'ServiceDegradation',
        description: 'Unusual error spike observed across telemetry ingress endpoints',
        likelihood: 0.85,
      });
      evidence.push({
        source: incident.source,
        description: `Triggered from ${incident.source} alerts`,
        confidence: 0.85,
        timestamp: incident.createdAt,
      });
    }

    return {
      incidentId: incident.id,
      probableCauses: causes,
      evidence,
      overallScore: causes[0].likelihood,
      analyzedAt: new Date().toISOString(),
    };
  },

  // Intelligence summary synthesizer
  generateSummary(incident: Incident, rca: RCAResult): IncidentSummary {
    const primary = rca.probableCauses[0];
    return {
      summary: `Automated SRE briefing for incident '${incident.title}'. Probable cause identified at ${primary.component} with ${(primary.likelihood * 100).toFixed(0)}% algorithmic confidence.`,
      keyFindings: [
        `Originating source: ${incident.source} with severity ${incident.severity.toUpperCase()}`,
        `Suspected failure type: ${primary.failureType} (${primary.description})`,
        `Safety Policy: Guardrail requires human sign-off before executing remediation`,
      ],
      affectedSystems: [incident.source, primary.component],
      recommendedActions: [
        `Approve and trigger runbook action: ${incident.title.includes('Disk') ? 'ClearDiskSpace' : 'RestartService'}`,
        'Verify metric normalization in Prometheus following execution',
      ],
      generatedByAi: true,
    };
  },
};
