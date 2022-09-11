package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// ConfigureTracing
// 1. Configures a tracer exporter (console, jaeger, ...)
// 2. Setup tracer provider
// 3. Globally register this provider and returns it.
func ConfigureTracing(_ context.Context, serviceName, serviceVersion string) (*sdktrace.TracerProvider, error) {
	exporter, err := makeExporter()
	if err != nil {
		return nil, err
	}

	provider, err := makeTracerProvider(serviceName, serviceVersion, exporter)
	if err != nil {
		return nil, err
	}

	// globally register the tracer exporter
	otel.SetTracerProvider(provider)
	return provider, nil
}

// makeExporter configures the Jaeger exporter
// Its need OTEL_EXPORTER_JAEGER_ENDPOINT environment variable to be set, or use default value if not defined
// See https://github.com/open-telemetry/opentelemetry-go/tree/v1.9.0/exporters/jaeger#environment-variables
func makeExporter() (sdktrace.SpanExporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint())
	if err != nil {
		return nil, fmt.Errorf("tracing: cannot init jaeger exporter: %w", err)
	}

	return exp, nil
}

func makeTracerProvider(serviceName, serviceVersion string, exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to configure tracer provider: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(r),
	), nil
}
