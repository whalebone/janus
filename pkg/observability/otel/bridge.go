package otel

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Config struct {
	ServiceName       string
	ServiceVersion    string
	ServiceInstanceID string
	SampleFraction    float64
	OTLPEndpoint      string
	OTLPProtocol      string // "grpc" or "http"
	OTLPInsecure      bool
}

// InitOTelBridge initializes the OpenTelemetry bridge for OpenCensus
func InitOTelBridge(cfg *Config) error {
	endpoint := cfg.OTLPEndpoint
	if endpoint == "" {
		endpoint = os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
		if endpoint == "" {
			endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		}
	}

	sampleFraction := cfg.SampleFraction
	if sampleFraction == 0 {
		sampleFraction = 1.0
	}

	var sampler sdktrace.Sampler
	if endpoint == "" {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(sampleFraction),
		)
	}

	serviceName := cfg.ServiceName
	if sn := os.Getenv("OTEL_SERVICE_NAME"); sn != "" {
		serviceName = sn
	}

	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
	}

	if cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(cfg.ServiceVersion))
	}

	if cfg.ServiceInstanceID != "" {
		attrs = append(attrs, semconv.ServiceInstanceID(cfg.ServiceInstanceID))
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(attrs...),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithHost(),
	)
	if err != nil {
		return err
	}

	if endpoint == "" {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sampler),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)

		opencensus.InstallTraceBridge(
			opencensus.WithTracerProvider(tp),
		)
		return nil
	}

	protocol := cfg.OTLPProtocol
	if protocol == "" {
		protocol = os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
		if protocol == "" {
			protocol = "grpc"
		}
	}

	var exporter sdktrace.SpanExporter
	if protocol == "http" || protocol == "http/protobuf" {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
		}
		if cfg.OTLPInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(context.Background(), opts...)
	} else {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint),
		}
		if cfg.OTLPInsecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(context.Background(), opts...)
	}

	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	opencensus.InstallTraceBridge(
		opencensus.WithTracerProvider(tp),
	)

	return nil
}
