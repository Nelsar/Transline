package otel

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Init инициализирует OpenTelemetry и возвращает shutdown-функцию
func Init(serviceName string) func(context.Context) error {
	ctx := context.Background()

	// OTLP exporter (в otel-collector)
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(getOtelEndpoint()),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create otlp exporter: %v", err)
	}

	// Resource — КТО генерирует трейсы
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("environment", getEnv()),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // важно для тестового задания
	)

	otel.SetTracerProvider(tp)

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return tp.Shutdown(ctx)
	}
}

func getOtelEndpoint() string {
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		return v
	}
	return "localhost:4317"
}

func getEnv() string {
	if v := os.Getenv("ENV"); v != "" {
		return v
	}
	return "local"
}
