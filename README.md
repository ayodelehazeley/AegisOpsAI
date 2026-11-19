# AegisOpsAI

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/AegisOpsAI/aegisops/actions/workflows/ci.yml/badge.svg)](.github/workflows/ci.yml)

> **Open-Source, Self-Hostable AI-Powered IT Operations and Incident Intelligence Platform**

AegisOpsAI is built for DevOps and SRE teams to detect, investigate, understand, and resolve infrastructure problems on their own terms. It acts as an intelligent incident and operational layer situated above your existing monitoring platforms (Prometheus, Alertmanager, OpenTelemetry, CloudWatch, Datadog), correlating alerts, deducing probable root causes, generating explainable summaries, and executing controlled remediations with mandatory human sign-off.

---

## 🎯 Core Principles

1. **Go-First Modular Monolith**: High performance, single-binary deployment, simple operational footprint. No microservice sprawl.
2. **Provider-Independent AI**: Seamlessly connect local offline LLMs (via **Ollama**) or cloud providers (**OpenAI**, Azure, Anthropic).
3. **Deterministic Core**: Core correlation, anomaly detection, policy enforcement, and audit logs continue working flawlessly even if AI providers are unreachable.
4. **Safety & Guardrails First**: Remediation actions strictly adhere to least-privilege RBAC, policy rules, and human-in-the-loop approvals before executing.
5. **Event Persistence**: Built on top of **NATS JetStream** and **PostgreSQL with pgvector** for reliable stream replays and semantic knowledge retrieval.

---

## 🏛 Architecture

```
                      ┌───────────────────────────────────────────────┐
                      │                 DATA SOURCES                  │
                      │  Prometheus · Alertmanager · OpenTelemetry   │
                      │       Docker · Kubernetes · Cloud APIs        │
                      └──────────────────────┬────────────────────────┘
                                             │ HTTP / Webhooks / gRPC
                                             ▼
                      ┌───────────────────────────────────────────────┐
                      │              GO INGESTION LAYER               │
                      │   Connectors · Normalizers · Common Schema    │
                      └──────────────────────┬────────────────────────┘
                                             │
                                             ▼
                      ┌───────────────────────────────────────────────┐
                      │             NATS JETSTREAM BROKER             │
                      │       EVENTS · INCIDENTS · ACTIONS STREAMS    │
                      └──────────────────────┬────────────────────────┘
                                             │
                                             ▼
                      ┌───────────────────────────────────────────────┐
                      │          INTELLIGENCE & RCA ENGINE            │
                      │   - Rule-based Anomaly Detection              │
                      │   - Alert Correlation & Grouping              │
                      │   - Deterministic Root Cause Analysis (RCA)   │
                      │   - Vector Knowledge Retrieval (pgvector)     │
                      │   - AI Provider (Ollama / OpenAI / Fallback)  │
                      └──────────────────────┬────────────────────────┘
                                             │
                                             ▼
                      ┌───────────────────────────────────────────────┐
                      │           POLICY & EXECUTION ENGINE           │
                      │   - Human Approval Workflow (RBAC)            │
                      │   - Remediation Runbooks & Rollback           │
                      │   - Immutable PostgreSQL Audit Trail          │
                      └──────────────────────┬────────────────────────┘
                                             │
                                             ▼
                      ┌───────────────────────────────────────────────┐
                      │          OPERATIONS REST API & UI             │
                      │   Chi Router · OpenAPI 3.1 · React Dashboard  │
                      └───────────────────────────────────────────────┘
```

---

## 🚀 Quick Start (Docker Compose)

Get AegisOpsAI running with PostgreSQL, pgvector, and NATS JetStream in 3 commands:

```bash
# 1. Clone the repository
git clone https://github.com/AegisOpsAI/aegisops.git
cd aegisops

# 2. Copy the sample environment configuration
cp configs/aegisops.yaml configs/local.yaml

# 3. Launch stack
docker compose up -d
```

Verify service readiness:
```bash
curl http://localhost:8080/healthz
```

---

## 📦 Project Structure

```
AegisOpsAI/
├── cmd/
│   └── aegisops/               # Application main entrypoint
├── internal/
│   ├── api/                    # REST API: handlers, routes, middleware
│   ├── audit/                  # Audit trail logging and queries
│   ├── auth/                   # JWT authentication, bcrypt passwords
│   ├── config/                 # Viper configuration loader
│   ├── connectors/             # External platform connectors (Prometheus, Alertmanager, Webhook)
│   ├── correlation/            # Event correlation & incident grouping
│   ├── database/               # PostgreSQL pool & repositories
│   ├── detection/              # Anomaly & threshold detection engine
│   ├── ingestion/              # NATS JetStream event publisher & consumer
│   ├── intelligence/           # AI summary generator & provider abstraction (Ollama, OpenAI)
│   ├── knowledge/              # Runbooks & vector search (pgvector)
│   ├── models/                 # Core domain models
│   ├── normalizer/             # Canonical event schema transformation
│   ├── policy/                 # RBAC and execution safety guardrails
│   ├── rca/                    # Deterministic Root Cause Analysis engine
│   ├── remediation/            # Propose, approve, and execute remediations
│   ├── scheduler/              # Background jobs & periodic health sweeps
│   └── telemetry/              # OpenTelemetry tracing and metrics
├── configs/                    # Default and local YAML configuration files
├── deployments/
│   ├── docker/                 # Database initialization and Docker assets
│   └── kubernetes/             # Helm charts and manifests
├── docs/                       # Architectural specs and guides
├── migrations/                 # Goose SQL database migrations
├── Makefile                    # Developer build, test, and run tasks
├── Dockerfile                  # Lightweight multi-stage Alpine container
└── docker-compose.yml          # Self-hosted single-node stack
```

---

## 🔌 Integrations

### Prometheus Alertmanager
Add AegisOpsAI as a webhook receiver in your `alertmanager.yml`:
```yaml
receivers:
  - name: 'aegisops'
    webhook_configs:
      - url: 'http://<aegisops-host>:8080/integrations/alertmanager'
        send_resolved: true
```

### Generic Webhooks
Send any structured operational event with HMAC-SHA256 signature support:
```bash
curl -X POST http://localhost:8080/integrations/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "source": "ci-runner",
    "title": "Pipeline Deployment Failed",
    "severity": "high",
    "description": "Kubernetes rollout failed due to CrashLoopBackOff"
  }'
```

---

## 🧠 AI Provider Setup

### Local LLM (Ollama)
Configure in `configs/aegisops.yaml` or via environment variables:
```yaml
ai:
  enabled: true
  provider: "ollama"
  model: "llama3"
  base_url: "http://localhost:11434"
```

### OpenAI / Compatible Endpoint (vLLM, LiteLLM)
```yaml
ai:
  enabled: true
  provider: "openai"
  model: "gpt-4o-mini"
  api_key: "sk-..."
```

---

## 🛡 Security & Incident Workflow

1. **Incident Triggered**: A critical alert arrives from Prometheus.
2. **Correlation & Grouping**: AegisOpsAI attaches duplicate and related service events to an open incident.
3. **Deterministic RCA**: Temporal cascading and resource exhaustion heuristics analyze evidence.
4. **AI Contextualization**: An explanation and proposed runbook action are generated.
5. **Policy Guardrail**: Action marked as `pending_approval`.
6. **Human Sign-Off**: SRE reviews evidence via the API/Dashboard and signs off.
7. **Execution & Verification**: Action is executed, and metric recovery is verified.
8. **Audit Trail**: Every step recorded immutably in PostgreSQL.

---

## 🤝 Contributing

We welcome contributions from DevOps, SRE, and Go engineers! Please check out [CONTRIBUTING.md](CONTRIBUTING.md) to get started.

---

## 📄 License

AegisOpsAI is distributed under the [Apache License 2.0](LICENSE).
