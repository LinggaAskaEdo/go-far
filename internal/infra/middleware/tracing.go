package middleware

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Tracing struct {
	tracer trace.Tracer
}

func InitTracing(serviceName string) *Tracing {
	return &Tracing{
		tracer: otel.Tracer(serviceName),
	}
}

func (t *Tracing) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			spanName := r.Method + " " + r.URL.Path
			ctx, span := t.tracer.Start(ctx, spanName,
				trace.WithAttributes(
					semconv.HTTPRoute(r.URL.Path),
					semconv.HTTPMethod(r.Method),
					semconv.URLPath(r.URL.Path),
				),
			)
			defer func() {
				span.SetAttributes(semconv.HTTPStatusCode(w.(interface{ StatusCode() int }).StatusCode()))
				span.End()
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (t *Tracing) StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

func (t *Tracing) TraceFunc(ctx context.Context, name string, fn func(ctx context.Context)) {
	ctx, span := t.tracer.Start(ctx, name)
	defer span.End()
	fn(ctx)
}

func (t *Tracing) WithSpan(ctx context.Context, operation string, fn func(context.Context) error) error {
	ctx, span := t.tracer.Start(ctx, operation)
	defer func() {
		if err := fn(ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()
	return fn(ctx)
}
