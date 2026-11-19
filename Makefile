# =============================================================================
# AegisOpsAI — Makefile
# =============================================================================

APP_NAME     := aegisops
CMD_DIR      := ./cmd/aegisops
BUILD_DIR    := ./bin
GO           := go
GOTEST       := $(GO) test
GOLINT       := golangci-lint
GOOSE        := goose

# Database connection (override via environment)
DB_DSN       ?= postgres://aegisops:aegisops_dev@localhost:5432/aegisops?sslmode=disable
MIGRATIONS   := ./migrations

.PHONY: all build run test test-integration lint migrate-up migrate-down \
        docker-build docker-up docker-down generate clean help

## help: Show this help message
help:
	@echo "AegisOpsAI — Available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

## build: Compile the Go binary
build:
	@echo "==> Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "==> Binary: $(BUILD_DIR)/$(APP_NAME)"

## run: Run the application via go run
run:
	$(GO) run $(CMD_DIR)

## test: Run all unit tests
test:
	$(GOTEST) -race -cover -count=1 ./...

## test-integration: Run integration tests (requires running services)
test-integration:
	$(GOTEST) -race -tags=integration -count=1 ./...

## lint: Run golangci-lint
lint:
	$(GOLINT) run ./...

## migrate-up: Apply all pending database migrations
migrate-up:
	$(GOOSE) -dir $(MIGRATIONS) postgres "$(DB_DSN)" up

## migrate-down: Roll back the last database migration
migrate-down:
	$(GOOSE) -dir $(MIGRATIONS) postgres "$(DB_DSN)" down

## docker-build: Build the Docker image
docker-build:
	docker build -t $(APP_NAME):latest .

## docker-up: Start the full stack with Docker Compose
docker-up:
	docker compose up -d --build

## docker-down: Stop the Docker Compose stack
docker-down:
	docker compose down

## generate: Run code generators (sqlc)
generate:
	sqlc generate

## clean: Remove build artifacts
clean:
	@echo "==> Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean -cache -testcache
	@echo "==> Done."

all: lint test build
