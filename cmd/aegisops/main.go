// Package main is the entrypoint for the AegisOpsAI platform.
// It bootstraps all subsystems — configuration, database, messaging,
// telemetry, background workers, and HTTP routing — then starts the server with graceful shutdown.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"

	"github.com/AegisOpsAI/aegisops/internal/api/handlers"
	"github.com/AegisOpsAI/aegisops/internal/api/routes"
	"github.com/AegisOpsAI/aegisops/internal/audit"
	"github.com/AegisOpsAI/aegisops/internal/auth"
	"github.com/AegisOpsAI/aegisops/internal/config"
	"github.com/AegisOpsAI/aegisops/internal/connectors/alertmanager"
	"github.com/AegisOpsAI/aegisops/internal/connectors/webhook"
	"github.com/AegisOpsAI/aegisops/internal/correlation"
	"github.com/AegisOpsAI/aegisops/internal/database"
	"github.com/AegisOpsAI/aegisops/internal/detection"
	"github.com/AegisOpsAI/aegisops/internal/ingestion"
	"github.com/AegisOpsAI/aegisops/internal/normalizer"
	"github.com/AegisOpsAI/aegisops/internal/remediation"
	"github.com/AegisOpsAI/aegisops/internal/scheduler"
	"github.com/AegisOpsAI/aegisops/internal/telemetry"
)

const (
	version = "0.1.0"
	banner  = `
    _    _____ ____ ___ ____   ___             _    ___
   / \  | ____/ ___|_ _/ ___| / _ \ _ __  ___ / \  |_ _|
  / _ \ |  _|| |  _ | |\___ \| | | | '_ \/ __/ _ \  | |
 / ___ \| |__| |_| || | ___) | |_| | |_) \__ \ ___ \ | |
/_/   \_\_____\____|___|____/ \___/| .__/|___/_/   \_\___|
                                   |_|
`
)

func main() {
	os.Exit(run())
}

func run() int {
	// ── Load configuration ──────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		return 1
	}

	// ── Set up structured logging ───────────────────────────────────────
	logLevel := slog.LevelInfo
	if cfg.Server.Mode == "development" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: cfg.Server.Mode == "development",
	}))
	slog.SetDefault(logger)

	// Print startup banner.
	fmt.Print(banner)
	slog.Info("starting AegisOpsAI",
		slog.String("version", version),
		slog.String("mode", cfg.Server.Mode),
		slog.String("addr", cfg.Server.Addr()),
	)

	// ── Context for graceful shutdown ───────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ── Initialize OpenTelemetry ────────────────────────────────────────
	var tp *telemetry.Provider
	if cfg.Telemetry.Enabled {
		tp, err = telemetry.InitTracer(ctx, cfg.Telemetry.ExporterEndpoint)
		if err != nil {
			slog.Error("failed to initialize telemetry", slog.String("error", err.Error()))
			return 1
		}
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if serr := tp.Shutdown(shutdownCtx); serr != nil {
				slog.Error("telemetry shutdown error", slog.String("error", serr.Error()))
			}
		}()
	} else {
		slog.Info("telemetry disabled, skipping OpenTelemetry initialization")
	}

	// ── Initialize database connection pool ─────────────────────────────
	dbPool, err := database.NewPool(ctx, cfg.Database.DSN(), cfg.Database.MaxConns)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		return 1
	}
	defer dbPool.Close()

	// Automatically ensure database schema & tables exist
	if err := database.EnsureSchema(ctx, dbPool); err != nil {
		slog.Warn("database schema check warning", slog.String("error", err.Error()))
	}

	// ── Connect to NATS JetStream (with fallback for standalone/cloud deployments) ──
	var nc *nats.Conn
	var js nats.JetStreamContext

	nc, err = nats.Connect(cfg.NATS.URL,
		nats.Name("aegisops"),
		nats.Timeout(3*time.Second),
		nats.RetryOnFailedConnect(false),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				slog.Warn("NATS disconnected", slog.String("error", err.Error()))
			}
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			slog.Info("NATS reconnected")
		}),
	)
	if err != nil {
		slog.Warn("NATS broker unavailable, running with direct storage pipeline", slog.String("url", cfg.NATS.URL), slog.String("error", err.Error()))
	} else {
		defer nc.Close()
		js, err = nc.JetStream()
		if err != nil {
			slog.Warn("failed to create JetStream context", slog.String("error", err.Error()))
		} else {
			slog.Info("NATS JetStream connected", slog.String("url", cfg.NATS.URL))
			if err := ingestion.SetupStreams(ctx, js, logger); err != nil {
				slog.Warn("JetStream streams setup warning", slog.String("error", err.Error()))
			}
		}
	}

	// ── Instantiate Repositories & Domain Services ──────────────────────
	repo := database.NewRepository(dbPool)
	normRegistry := normalizer.NewRegistry()
	ingester := ingestion.NewEventIngester(js, js, repo, normRegistry, logger)

	authSvc := auth.NewService(auth.Config{
		JWTSecret:       cfg.Auth.JWTSecret,
		AccessTokenTTL:  cfg.Auth.TokenExpiry,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "aegisops",
	})

	remSvc := remediation.NewService(repo, logger)
	auditSvc := audit.NewService(dbPool, logger)
	correlationEngine := correlation.NewEngine(repo, 30*time.Minute, logger)
	detectionEngine := detection.NewEngine(logger)
	_ = detectionEngine

	// ── Start Event Consumers & Background Workers ──────────────────────
	if _, err := ingester.StartConsumer(ctx); err != nil {
		slog.Warn("event consumer failed to start on JetStream", slog.String("error", err.Error()))
	}

	sched := scheduler.NewScheduler(logger)
	sched.RegisterJob("incident-correlation-sweep", 1*time.Minute, func(jobCtx context.Context) error {
		logger.Debug("running scheduled correlation heartbeat sweep")
		return nil
	})
	sched.Start(ctx)
	defer sched.Stop()

	// ── Initialize HTTP Handlers ────────────────────────────────────────
	healthH := handlers.NewHealthHandler(dbPool, nc, logger)
	authH := handlers.NewAuthHandler(dbPool, authSvc, logger)
	incidentsH := handlers.NewIncidentHandler(dbPool, logger)
	eventsH := handlers.NewEventHandler(dbPool, ingester, logger)
	remH := handlers.NewRemediationHandler(remSvc, logger)
	auditH := handlers.NewAuditHandler(auditSvc, logger)

	// Ingestion connectors
	alertmanagerHandler := alertmanager.NewHandler(func(event *normalizer.NormalizedEvent) error {
		_, err := correlationEngine.CorrelateEvent(ctx, event)
		return err
	}, logger)

	webhookHandler := webhook.NewHandler(webhook.Config{DefaultSource: "webhook"}, func(event *normalizer.NormalizedEvent) error {
		_, err := correlationEngine.CorrelateEvent(ctx, event)
		return err
	}, logger)

	// ── Build HTTP router ───────────────────────────────────────────────
	r := chi.NewRouter()

	// Global middleware stack.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(slogMiddleware)
	r.Use(middleware.Recoverer)

	if cfg.Telemetry.Enabled {
		r.Use(telemetry.HTTPMiddleware)
	}

	// Register all module routes.
	routes.Register(r, &routes.Handlers{
		Health:       healthH,
		Auth:         authH,
		Incidents:    incidentsH,
		Events:       eventsH,
		Remediations: remH,
		Audit:        auditH,
		Alertmanager: alertmanagerHandler,
		Webhook:      webhookHandler,
	}, authSvc, logger)

	// Serve static SPA dashboard if directory exists
	staticDirs := []string{"./web/dashboard/dist", "/app/web/dashboard/dist"}
	for _, staticDir := range staticDirs {
		if _, err := os.Stat(staticDir); err == nil {
			fs := http.FileServer(http.Dir(staticDir))
			r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
				if strings.HasPrefix(req.URL.Path, "/api") || strings.HasPrefix(req.URL.Path, "/integrations") || req.URL.Path == "/healthz" {
					http.NotFound(w, req)
					return
				}
				fPath := filepath.Join(staticDir, req.URL.Path)
				if _, err := os.Stat(fPath); os.IsNotExist(err) {
					http.ServeFile(w, req, filepath.Join(staticDir, "index.html"))
					return
				}
				fs.ServeHTTP(w, req)
			})
			slog.Info("serving dashboard static assets", slog.String("dir", staticDir))
			break
		}
	}

	// ── Start HTTP server ───────────────────────────────────────────────
	srv := &http.Server{
		Addr:              cfg.Server.Addr(),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", slog.String("addr", srv.Addr))
		if serr := srv.ListenAndServe(); serr != nil && serr != http.ErrServerClosed {
			errCh <- serr
		}
		close(errCh)
	}()

	// ── Wait for shutdown signal ────────────────────────────────────────
	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining connections...")
	case err := <-errCh:
		slog.Error("server error, initiating shutdown", slog.String("error", err.Error()))
	}

	// Graceful shutdown with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", slog.String("error", err.Error()))
		return 1
	}

	slog.Info("AegisOpsAI shut down gracefully")
	return 0
}

func slogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			slog.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
