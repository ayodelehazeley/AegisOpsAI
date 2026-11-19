-- +goose Up
-- =============================================================================
-- AegisOpsAI — Initial Schema
-- =============================================================================

-- ---------------------------------------------------------------------------
-- users
-- ---------------------------------------------------------------------------
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name  TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT 'viewer'
                  CHECK (role IN ('admin', 'operator', 'viewer')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role  ON users (role);

-- ---------------------------------------------------------------------------
-- incidents
-- ---------------------------------------------------------------------------
CREATE TABLE incidents (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    severity    TEXT NOT NULL DEFAULT 'info'
                CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    status      TEXT NOT NULL DEFAULT 'open'
                CHECK (status IN ('open', 'acknowledged', 'investigating', 'mitigated', 'resolved', 'closed')),
    source      TEXT NOT NULL DEFAULT '',
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_incidents_severity    ON incidents (severity);
CREATE INDEX idx_incidents_status      ON incidents (status);
CREATE INDEX idx_incidents_assignee_id ON incidents (assignee_id);
CREATE INDEX idx_incidents_created_at  ON incidents (created_at);

-- ---------------------------------------------------------------------------
-- events
-- ---------------------------------------------------------------------------
CREATE TABLE events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source      TEXT NOT NULL,
    type        TEXT NOT NULL,
    severity    TEXT NOT NULL DEFAULT 'info'
                CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    payload     JSONB NOT NULL DEFAULT '{}',
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT now(),
    incident_id UUID REFERENCES incidents(id) ON DELETE SET NULL,
    normalized  BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_source      ON events (source);
CREATE INDEX idx_events_type        ON events (type);
CREATE INDEX idx_events_severity    ON events (severity);
CREATE INDEX idx_events_timestamp   ON events (timestamp);
CREATE INDEX idx_events_incident_id ON events (incident_id);
CREATE INDEX idx_events_payload     ON events USING GIN (payload);

-- ---------------------------------------------------------------------------
-- alerts
-- ---------------------------------------------------------------------------
CREATE TABLE alerts (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source      TEXT NOT NULL,
    name        TEXT NOT NULL,
    labels      JSONB NOT NULL DEFAULT '{}',
    severity    TEXT NOT NULL DEFAULT 'info'
                CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    status      TEXT NOT NULL DEFAULT 'firing'
                CHECK (status IN ('firing', 'resolved', 'silenced')),
    fired_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ,
    event_id    UUID REFERENCES events(id) ON DELETE SET NULL
);

CREATE INDEX idx_alerts_source   ON alerts (source);
CREATE INDEX idx_alerts_severity ON alerts (severity);
CREATE INDEX idx_alerts_status   ON alerts (status);
CREATE INDEX idx_alerts_fired_at ON alerts (fired_at);
CREATE INDEX idx_alerts_event_id ON alerts (event_id);
CREATE INDEX idx_alerts_labels   ON alerts USING GIN (labels);

-- ---------------------------------------------------------------------------
-- audit_logs
-- ---------------------------------------------------------------------------
CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    action        TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id   UUID,
    details       JSONB NOT NULL DEFAULT '{}',
    timestamp     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user_id       ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_action        ON audit_logs (action);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs (resource_type);
CREATE INDEX idx_audit_logs_timestamp     ON audit_logs (timestamp);
CREATE INDEX idx_audit_logs_details       ON audit_logs USING GIN (details);

-- ---------------------------------------------------------------------------
-- remediations
-- ---------------------------------------------------------------------------
CREATE TABLE remediations (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    description TEXT NOT NULL DEFAULT '',
    type        TEXT NOT NULL DEFAULT 'manual'
                CHECK (type IN ('manual', 'automated', 'ai_suggested')),
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'approved', 'executing', 'completed', 'failed', 'rejected')),
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    executed_at TIMESTAMPTZ,
    result      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_remediations_incident_id ON remediations (incident_id);
CREATE INDEX idx_remediations_status      ON remediations (status);
CREATE INDEX idx_remediations_type        ON remediations (type);

-- ---------------------------------------------------------------------------
-- runbooks
-- ---------------------------------------------------------------------------
CREATE TABLE runbooks (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    steps       JSONB NOT NULL DEFAULT '[]',
    tags        TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_runbooks_tags  ON runbooks USING GIN (tags);
CREATE INDEX idx_runbooks_steps ON runbooks USING GIN (steps);

-- ---------------------------------------------------------------------------
-- knowledge_embeddings
-- ---------------------------------------------------------------------------
CREATE TABLE knowledge_embeddings (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    incident_id UUID REFERENCES incidents(id) ON DELETE CASCADE,
    content     TEXT NOT NULL,
    embedding   vector(1536) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_knowledge_embeddings_incident_id ON knowledge_embeddings (incident_id);
CREATE INDEX idx_knowledge_embeddings_vector       ON knowledge_embeddings
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);


-- +goose Down
DROP TABLE IF EXISTS knowledge_embeddings;
DROP TABLE IF EXISTS runbooks;
DROP TABLE IF EXISTS remediations;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS users;
