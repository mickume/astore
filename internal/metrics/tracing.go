package metrics

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"zotregistry.io/zot/pkg/log"
)

const (
	tracerName = "zot-artifact-store"
)

// TracingProvider manages OpenTelemetry tracing
type TracingProvider struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	logger   log.Logger
	enabled  bool
}

// NewTracingProvider creates a new tracing provider
func NewTracingProvider(ctx context.Context, endpoint string, serviceName string, logger log.Logger) (*TracingProvider, error) {
	if endpoint == "" {
		// Tracing disabled
		return &TracingProvider{
			enabled: false,
			logger:  logger,
		}, nil
	}

	// Create OTLP exporter
	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(), // Use TLS in production
		),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(provider)

	tracer := provider.Tracer(tracerName)

	logger.Info().
		Str("endpoint", endpoint).
		Str("service", serviceName).
		Msg("OpenTelemetry tracing initialized")

	return &TracingProvider{
		tracer:   tracer,
		provider: provider,
		logger:   logger,
		enabled:  true,
	}, nil
}

// Shutdown shuts down the tracing provider
func (t *TracingProvider) Shutdown(ctx context.Context) error {
	if !t.enabled || t.provider == nil {
		return nil
	}

	return t.provider.Shutdown(ctx)
}

// StartSpan starts a new trace span
func (t *TracingProvider) StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}

	return t.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// TraceArtifactOperation creates a span for artifact operations
func (t *TracingProvider) TraceArtifactOperation(ctx context.Context, operation, bucket, key string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "artifact."+operation,
		attribute.String("artifact.operation", operation),
		attribute.String("artifact.bucket", bucket),
		attribute.String("artifact.key", key),
	)
}

// TraceSupplyChainOperation creates a span for supply chain operations
func (t *TracingProvider) TraceSupplyChainOperation(ctx context.Context, operation, artifactID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "supplychain."+operation,
		attribute.String("supplychain.operation", operation),
		attribute.String("supplychain.artifact_id", artifactID),
	)
}

// TraceAuthOperation creates a span for authentication/authorization operations
func (t *TracingProvider) TraceAuthOperation(ctx context.Context, operation string, userID string) (context.Context, trace.Span) {
	return t.StartSpan(ctx, "auth."+operation,
		attribute.String("auth.operation", operation),
		attribute.String("auth.user_id", userID),
	)
}

// AddSpanAttributes adds attributes to the current span
func (t *TracingProvider) AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if !t.enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error in the current span
func (t *TracingProvider) RecordError(ctx context.Context, err error) {
	if !t.enabled || err == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
