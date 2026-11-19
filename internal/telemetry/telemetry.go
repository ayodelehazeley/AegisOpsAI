// Package telemetry provides OpenTelemetry initialization for distributed
// tracing and metrics collection in AegisOpsAI. It configures OTLP exporters
// and provides HTTP middleware for automatic span creation.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName    = "aegisops"
	serviceVersion = "0.1.0"
)

// Provider wraps the OpenTelemetry TracerProvider and provides
// lifecycle management for the telemetry subsystem.
type Provider struct {
	tp *sdktrace.TracerProvider
}

// InitTracer initializes an OTLP gRPC trace exporter and configures
// the global OpenTelemetry TracerProvider. Returns a Provider that
// must be shut down on application exit.
func InitTracer(ctx context.Context, endpoint string) (*Provider, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
		resource.WithHost(),
		resource.WithOS(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating telemetry resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global providers and propagators.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("opentelemetry tracer initialized",
		slog.String("endpoint", endpoint),
		slog.String("service", serviceName),
	)

	return &Provider{tp: tp}, nil
}

// Shutdown gracefully flushes and shuts down the TracerProvider.
// It should be called during application shutdown to ensure all
// pending spans are exported.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.tp == nil {
		return nil
	}

	slog.Info("shutting down opentelemetry tracer")
	if err := p.tp.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down tracer provider: %w", err)
	}
	return nil
}

// Tracer returns a named tracer from the global TracerProvider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// HTTPMiddleware returns Chi-compatible middleware that creates a span
// for each incoming HTTP request with standard HTTP semantic conventions.
func HTTPMiddleware(next http.Handler) http.Handler {
	tracer := otel.Tracer("aegisops.http")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract any propagated context from incoming headers.
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.URLPath(r.URL.Path),
				semconv.ServerAddress(r.Host),
				semconv.UserAgentOriginal(r.UserAgent()),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Wrap the response writer to capture the status code.
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sw, r.WithContext(ctx))

		span.SetAttributes(
			semconv.HTTPResponseStatusCode(sw.statusCode),
		)
	})
}

// statusWriter wraps http.ResponseWriter to capture the response status code.
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and delegates to the wrapped writer.
func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}
