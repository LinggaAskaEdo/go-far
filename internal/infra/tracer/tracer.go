package tracer

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// TracerOptions holds tracer configuration
type TracerOptions struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

// Tracer defines the tracer interface
type Tracer interface {
	Stop()
}

type tracerImpl struct {
	log      *zerolog.Logger
	provider *sdktrace.TracerProvider
}

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(log *zerolog.Logger, opt *TracerOptions) Tracer {
	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("go-far-app"),
			semconv.ServiceVersion("1.10.0"),
		),
	)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to create resource for tracer")
		return nil
	}

	// Use endpoint from config or default
	endpoint := opt.Endpoint
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to create OTLP exporter for tracer")
		return nil
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	log.Print("Tracer initialized successfully")

	return &tracerImpl{
		log:      log,
		provider: tp,
	}
}

// Stop shuts down the tracer and flushes any remaining spans
func (t *tracerImpl) Stop() {
	if t.provider == nil {
		t.log.Print("Tracer provider is nil, nothing to shut down...")
	}

	t.log.Print("Shutting down tracer...")
	if err := t.provider.Shutdown(context.Background()); err != nil {
		t.log.Printf("Error shutting down tracer: %v", err)
	}

	t.log.Print("Tracer shutdown complete...")
}
