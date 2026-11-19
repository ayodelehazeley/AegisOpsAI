export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'info';

export type IncidentStatus = 'open' | 'investigating' | 'identified' | 'monitoring' | 'resolved';

export type RemediationStatus = 'pending' | 'approved' | 'executing' | 'completed' | 'failed' | 'rejected';

export interface Incident {
  id: string;
  title: string;
  description: string;
  severity: Severity;
  status: IncidentStatus;
  source: string;
  createdAt: string;
  updatedAt: string;
  resolvedAt?: string | null;
  assigneeId?: string | null;
}

export interface OperationalEvent {
  id: string;
  source: string;
  type: string;
  severity: Severity;
  payload: Record<string, any>;
  timestamp: string;
  incidentId?: string | null;
  normalized: boolean;
  createdAt: string;
}

export interface Remediation {
  id: string;
  incidentId: string;
  description: string;
  type: string;
  status: RemediationStatus;
  approvedBy?: string | null;
  executedAt?: string | null;
  result?: string | null;
  createdAt: string;
}

export interface AuditLog {
  id: string;
  userId?: string | null;
  action: string;
  resourceType: string;
  resourceId?: string | null;
  details: Record<string, any>;
  timestamp: string;
}

export interface ComponentStatus {
  name: string;
  status: string;
  message?: string;
}

export interface SystemHealth {
  status: string;
  components: ComponentStatus[];
}

export interface RCAEvidence {
  source: string;
  description: string;
  confidence: number;
  timestamp: string;
}

export interface RCACause {
  component: string;
  failureType: string;
  description: string;
  likelihood: number;
  evidenceIds?: string[];
}

export interface RCAResult {
  incidentId: string;
  probableCauses: RCACause[];
  evidence: RCAEvidence[];
  overallScore: number;
  analyzedAt: string;
}

export interface IncidentSummary {
  summary: string;
  keyFindings: string[];
  affectedSystems: string[];
  recommendedActions: string[];
  generatedByAi: boolean;
}
