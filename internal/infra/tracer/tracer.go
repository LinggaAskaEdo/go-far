package tracer

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// TracerOptions holds tracer configuration
type TracerOptions struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Protocol string `yaml:"protocol"`
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
}

// Tracer defines the tracer interface
type Tracer interface {
	Stop()
	GetServiceName() string
	ForceFlush(context.Context) error
}

type tracerImpl struct {
	log         *zerolog.Logger
	provider    *sdktrace.TracerProvider
	serviceName string
}

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(log *zerolog.Logger, opt *TracerOptions) Tracer {
	if !opt.Enabled {
		return &tracerImpl{log: log, serviceName: opt.Name}
	}

	ctx := context.Background()

	serviceName := opt.Name
	if serviceName == "" {
		serviceName = "go-far-app"
	}

	serviceVersion := opt.Version
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create resource for tracer")
		return nil
	}

	endpoint := opt.Endpoint
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	protocol := opt.Protocol
	if protocol == "" {
		protocol = "grpc"
	}

	tp := createTracerProvider(ctx, log, endpoint, protocol, res)
	if tp == nil {
		return nil
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Info().Str("endpoint", endpoint).Str("protocol", protocol).Msg("Tracer initialized successfully")

	return &tracerImpl{
		log:         log,
		provider:    tp,
		serviceName: serviceName,
	}
}

func createTracerProvider(ctx context.Context, log *zerolog.Logger, endpoint, protocol string, res *resource.Resource) *sdktrace.TracerProvider {
	if protocol == "http" {
		return createHTTPProvider(ctx, log, endpoint, res)
	}
	return createGRPCProvider(ctx, log, endpoint, res)
}

func createHTTPProvider(ctx context.Context, log *zerolog.Logger, endpoint string, res *resource.Resource) *sdktrace.TracerProvider {
	log.Info().Str("endpoint", endpoint).Msg("Creating HTTP OTLP exporter")

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP OTLP exporter for tracer")
		return nil
	}

	log.Info().Msg("HTTP OTLP exporter created, creating TracerProvider")

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
}

func createGRPCProvider(ctx context.Context, log *zerolog.Logger, endpoint string, res *resource.Resource) *sdktrace.TracerProvider {
	log.Info().Str("endpoint", endpoint).Msg("Creating gRPC OTLP exporter")

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create OTLP exporter for tracer")
		return nil
	}

	log.Info().Msg("gRPC OTLP exporter created, creating TracerProvider")

	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
}

// Stop shuts down the tracer and flushes any remaining spans
func (t *tracerImpl) Stop() {
	if t.provider == nil {
		return
	}

	t.log.Info().Msg("Shutting down tracer...")
	if err := t.provider.Shutdown(context.Background()); err != nil {
		t.log.Error().Err(err).Msg("Error shutting down tracer")
		return
	}

	t.log.Info().Msg("Tracer shutdown complete")
}

func (t *tracerImpl) GetServiceName() string {
	return t.serviceName
}

func (t *tracerImpl) ForceFlush(ctx context.Context) error {
	if t.provider == nil {
		return nil
	}
	return t.provider.ForceFlush(ctx)
}
