# Architecture Overview

## System Overview

AegisOpsAI is built as a **modular monolith** — a single deployable binary composed of well-isolated internal modules that communicate through defined interfaces and an event bus. This architecture provides the development simplicity of a monolith with the modularity benefits of microservices, making it easy to self-host while maintaining clean separation of concerns.

## Module Descriptions

| Module        | Package                      | Responsibility                                                    |
|---------------|------------------------------|-------------------------------------------------------------------|
| **Auth**      | `internal/auth`              | JWT authentication, RBAC, user management                         |
| **Ingestion** | `internal/ingestion`         | Receive events from external sources (Prometheus, webhooks, etc.) |
| **Events**    | `internal/events`            | Event normalization, storage, and correlation                     |
| **Incidents** | `internal/incidents`         | Incident lifecycle management, severity tracking                  |
| **Alerts**    | `internal/alerts`            | Alert processing, deduplication, grouping                         |
| **AI Engine** | `internal/ai`                | LLM integration (Ollama/OpenAI), RCA, summarization              |
| **Knowledge** | `internal/knowledge`         | Embedding storage, similarity search (pgvector)                   |
| **Runbooks**  | `internal/runbooks`          | Runbook CRUD, step execution templates                            |
| **Remediation** | `internal/remediation`     | Remediation suggestions, approval workflows, execution            |
| **Audit**     | `internal/audit`             | Audit logging for all state-changing operations                   |
| **Platform**  | `internal/platform`          | Shared infrastructure: database, NATS, config, middleware         |

## Data Flow

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          External Sources                                │
│   Prometheus/Alertmanager  │  OpenTelemetry  │  Webhooks  │  API Calls   │
└──────────────┬───────────────────┬──────────────────┬────────────────────┘
               │                   │                  │
               ▼                   ▼                  ▼
        ┌─────────────────────────────────────────────────────┐
        │                 Ingestion Module                     │
        │   Validates, normalizes, and routes incoming data    │
        └──────────────────────┬──────────────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │   NATS JetStream    │  ◄── Event Bus
                    │   (Message Broker)  │
                    └──┬──────┬──────┬────┘
                       │      │      │
           ┌───────────┘      │      └───────────┐
           ▼                  ▼                  ▼
   ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐
   │ Events Module │  │ Alerts Module│  │ Incidents Module  │
   │  Correlate &  │  │  Deduplicate │  │  Create/Manage   │
   │  Store        │  │  & Group     │  │  Incidents       │
   └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘
          │                  │                   │
          └──────────┬───────┘                   │
                     ▼                           │
              ┌─────────────┐                    │
              │  AI Engine  │◄───────────────────┘
              │  RCA, NLP   │
              │  Summarize  │
              └──────┬──────┘
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
   ┌─────────────┐    ┌──────────────────┐
   │  Knowledge  │    │   Remediation    │
   │  Embeddings │    │   Suggestions &  │
   │  & Search   │    │   Execution      │
   └─────────────┘    └──────────────────┘

                    ┌─────────────┐
                    │ PostgreSQL  │  ◄── Persistent Storage
                    │ + pgvector  │
                    └─────────────┘
```

## Technology Choices

| Technology          | Purpose                    | Rationale                                                             |
|---------------------|----------------------------|-----------------------------------------------------------------------|
| **Go 1.23**         | Application language       | Performance, concurrency, single binary deployment, strong stdlib     |
| **Chi Router**      | HTTP routing               | Lightweight, idiomatic, stdlib-compatible, composable middleware      |
| **PostgreSQL 16**   | Primary database           | Mature, JSONB support, extensible (pgvector), strong ecosystem       |
| **pgvector**        | Vector similarity search   | Native Postgres extension, no separate vector DB needed               |
| **pgx/v5**          | Database driver             | Pure Go, high performance, native PostgreSQL type support             |
| **NATS JetStream**  | Message broker             | Lightweight, at-least-once delivery, single binary, easy to operate   |
| **Viper**           | Configuration              | Multi-source config (file, env, flags), widely adopted                |
| **log/slog**        | Structured logging         | Stdlib, zero-dependency, JSON output, contextual fields               |
| **OpenTelemetry**   | Observability              | Vendor-neutral traces and metrics, industry standard                  |
| **JWT**             | Authentication             | Stateless, scalable, well-understood                                  |
| **goose**           | Database migrations        | SQL-based, simple, Go-native                                          |
| **Docker**          | Containerization           | Universal deployment, reproducible environments                       |

## Design Principles

1. **Self-hostable first** — Single binary + Docker Compose; no cloud dependencies required.
2. **AI-augmented, human-in-the-loop** — AI suggests; humans approve and execute.
3. **Event-driven internals** — Modules communicate via NATS JetStream for loose coupling.
4. **Observable by default** — Structured logging and OpenTelemetry built into every module.
5. **Secure by design** — RBAC, audit logging, and approval workflows are core, not afterthoughts.
