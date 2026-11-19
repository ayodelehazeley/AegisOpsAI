# Quick Start Guide

Get AegisOpsAI running locally in under 5 minutes.

## Prerequisites

| Requirement         | Minimum Version |
|---------------------|-----------------|
| Docker              | 24.0+           |
| Docker Compose      | 2.20+           |
| curl (for testing)  | any             |

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/AegisOpsAI/aegisops.git
cd aegisops
```

### 2. Start the Stack

```bash
docker compose up -d
```

This starts three services:
- **AegisOpsAI** — API server on [http://localhost:8080](http://localhost:8080)
- **PostgreSQL** — Database on port 5432 (with pgvector)
- **NATS** — Message broker on port 4222 (monitoring on [http://localhost:8222](http://localhost:8222))

### 3. Verify Health

```bash
curl http://localhost:8080/healthz
```

Expected response:
```json
{
  "status": "ok",
  "version": "dev"
}
```

## First Run Walkthrough

### Create Your First Admin User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecureP@ss123",
    "display_name": "Admin User"
  }'
```

### Authenticate

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecureP@ss123"
  }'
```

Save the `token` from the response for subsequent requests:

```bash
export TOKEN="<your-jwt-token>"
```

### Create an Incident

```bash
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "High CPU on web-server-01",
    "description": "CPU usage exceeded 95% for 10 minutes",
    "severity": "high",
    "source": "manual"
  }'
```

### List Incidents

```bash
curl http://localhost:8080/api/v1/incidents \
  -H "Authorization: Bearer $TOKEN"
```

## Connecting Prometheus Alertmanager

Configure Alertmanager to send alerts to AegisOpsAI by adding a webhook receiver:

```yaml
# alertmanager.yml
receivers:
  - name: aegisops
    webhook_configs:
      - url: http://<aegisops-host>:8080/api/v1/webhooks/alertmanager
        send_resolved: true

route:
  receiver: aegisops
```

Reload Alertmanager and verify alerts appear in AegisOpsAI:

```bash
curl http://localhost:8080/api/v1/alerts \
  -H "Authorization: Bearer $TOKEN"
```

## Configuration

AegisOpsAI is configured via environment variables or a config file at `configs/config.yaml`.

Key environment variables:

| Variable                       | Default              | Description                |
|--------------------------------|----------------------|----------------------------|
| `AEGISOPS_SERVER_PORT`         | `8080`               | HTTP server port           |
| `AEGISOPS_DATABASE_HOST`       | `localhost`          | PostgreSQL host            |
| `AEGISOPS_DATABASE_PORT`       | `5432`               | PostgreSQL port            |
| `AEGISOPS_DATABASE_USER`       | `aegisops`           | Database user              |
| `AEGISOPS_DATABASE_PASSWORD`   | `aegisops_dev`       | Database password          |
| `AEGISOPS_DATABASE_NAME`       | `aegisops`           | Database name              |
| `AEGISOPS_NATS_URL`            | `nats://localhost:4222` | NATS server URL         |
| `AEGISOPS_AUTH_JWT_SECRET`     | *(required)*         | JWT signing secret         |
| `AEGISOPS_AI_PROVIDER`         | `ollama`             | AI provider (ollama/openai)|
| `AEGISOPS_AI_OLLAMA_URL`       | `http://localhost:11434` | Ollama API URL         |
| `AEGISOPS_AI_OPENAI_API_KEY`   | *(none)*             | OpenAI API key             |
| `AEGISOPS_LOG_LEVEL`           | `info`               | Log level                  |

## Stopping the Stack

```bash
docker compose down
```

To also remove stored data:

```bash
docker compose down -v
```

## Next Steps

- Read the [Architecture Overview](../architecture/overview.md) to understand system design.
- Check the [API documentation](http://localhost:8080/docs) (available when the server is running).
- Set up [AI providers](../../README.md#ai-provider-setup) for intelligent incident analysis.
