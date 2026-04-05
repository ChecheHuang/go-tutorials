// Package telemetry 初始化 OpenTelemetry 追蹤（第 35 課）
package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// InitTracer 初始化 TracerProvider，回傳 shutdown 函式
func InitTracer(serviceName string) (trace.Tracer, func()) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		slog.Error("建立 trace exporter 失敗", "error", err)
		return otel.Tracer(serviceName), func() {}
	}

	res, _ := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer(serviceName)

	shutdown := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("關閉 TracerProvider 失敗", "error", err)
		}
	}

	slog.Info("OpenTelemetry 追蹤已初始化", "service", serviceName)
	return tracer, shutdown
}
