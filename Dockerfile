# =============================================================================
# AegisOpsAI — Multi-stage Docker Build
# =============================================================================

# ---------------------------------------------------------------------------
# Stage 1: Frontend UI Build
# ---------------------------------------------------------------------------
FROM node:20-alpine AS ui-builder

WORKDIR /web
COPY web/dashboard/package*.json ./
RUN npm ci || npm install
COPY web/dashboard/ ./
RUN npm run build

# ---------------------------------------------------------------------------
# Stage 2: Go Backend Build
# ---------------------------------------------------------------------------
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Cache dependency downloads
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /bin/aegisops ./cmd/aegisops

# ---------------------------------------------------------------------------
# Stage 3: Runtime
# ---------------------------------------------------------------------------
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -S aegisops \
    && adduser -S -G aegisops aegisops

WORKDIR /app

# Copy binary, config, and static dashboard assets
COPY --from=builder /bin/aegisops /app/aegisops
COPY --from=builder /src/configs/ /app/configs/
COPY --from=ui-builder /web/dist/ /app/web/dashboard/dist/

# Own everything by the non-root user
RUN chown -R aegisops:aegisops /app

USER aegisops

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/aegisops"]
