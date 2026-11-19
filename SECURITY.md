# Security Policy

## Reporting Vulnerabilities

If you discover a security vulnerability in AegisOpsAI, **please do not open a public issue.**

Instead, report it responsibly:

1. **Email**: Send details to **security@aegisopsai.dev**
2. **Include**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will acknowledge your report within **48 hours** and provide an initial assessment within **5 business days**.

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| latest  | ✅ Active support   |
| < 1.0   | ⚠️ Best effort      |

## Security Features

AegisOpsAI includes the following security features:

### Authentication & Authorization

- **JWT-based authentication** with configurable token expiry and refresh tokens.
- **Role-Based Access Control (RBAC)** with three built-in roles:
  - `admin` — Full system access
  - `operator` — Manage incidents, execute remediations
  - `viewer` — Read-only access to dashboards and incidents
- **Password hashing** using bcrypt with configurable cost factor.

### Audit & Accountability

- **Comprehensive audit logging** — All state-changing operations are recorded with user ID, action, resource, and timestamp.
- **Audit log immutability** — Audit records are append-only.

### Remediation Safety

- **Approval workflows** — AI-suggested and automated remediations require explicit human approval before execution.
- **Execution logging** — All remediation results are recorded for post-incident review.

### API Security

- **Input validation** on all API endpoints.
- **Rate limiting** to prevent abuse.
- **CORS** configuration for frontend integration.
- **Request ID** tracking for traceability.

### Infrastructure

- **Non-root container** — The Docker image runs as a non-root user.
- **TLS support** — Configurable TLS for the HTTP server.
- **Secret management** — Sensitive configuration via environment variables (never logged).

## Security Best Practices for Deployers

1. **Change default secrets** — Always override `AEGISOPS_AUTH_JWT_SECRET` in production.
2. **Enable TLS** — Run behind a reverse proxy with TLS termination or enable built-in TLS.
3. **Restrict database access** — Use a dedicated database user with minimal privileges.
4. **Network segmentation** — Place NATS and PostgreSQL on internal networks only.
5. **Regular updates** — Keep AegisOpsAI and its dependencies up to date.
6. **Monitor audit logs** — Review audit logs regularly for suspicious activity.

## Disclosure Policy

- We follow **coordinated disclosure**: fixes are developed privately and released before public disclosure.
- Credit is given to reporters (unless they prefer anonymity).
- CVEs are requested for significant vulnerabilities.
