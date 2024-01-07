package metric

import (
	"context"
	"fmt"
	"github.com/delfimarime/hermes/services/smsc/pkg/common"
	"github.com/delfimarime/hermes/services/smsc/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

func Configure(ctx context.Context, cfg config.Configuration) {
	if provider := getTracerProviderFrom(ctx, cfg.OpenTelemetry); provider != nil {
		otel.SetTracerProvider(provider)
	}
	if provider := getMeterProviderFrom(ctx, cfg.OpenTelemetry); provider != nil {
		otel.SetMeterProvider(provider)
	}
}

func getMeterProviderFrom(ctx context.Context, cfg config.OpenTelemetry) *metric.MeterProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		zap.L().Error("Cannot configure OpenTelemetry metric provider",
			zap.String("method", cfg.Metrics.ExportMethod),
			zap.String("service", cfg.ServiceName),
			zap.String("endpoint", cfg.Metrics.Endpoint),
			zap.Error(err),
		)
		return nil
	}
	exp, err := getMetricExporterFrom(ctx, cfg)
	if err != nil {
		zap.L().Error("Cannot configure OpenTelemetry metric provider",
			zap.String("method", cfg.Metrics.ExportMethod),
			zap.String("service", cfg.ServiceName),
			zap.String("endpoint", cfg.Metrics.Endpoint),
			zap.Error(err),
		)
		return nil
	}
	if exp == nil {
		zap.L().Warn("Cannot configure OpenTelemetry metric provider, method not specified")
		return nil
	}
	metricReaderOpts := make([]metric.PeriodicReaderOption, 0)
	if cfg.Metrics.Reader != nil {
		metricReaderOpts = append(metricReaderOpts,
			metric.WithInterval(common.MillisToDuration(cfg.Metrics.Reader.Period)))
	}
	return metric.NewMeterProvider(
		metric.WithResource(r),
		metric.WithReader(metric.NewPeriodicReader(exp, metricReaderOpts...)),
	)
}

func getMetricExporterFrom(ctx context.Context, cfg config.OpenTelemetry) (metric.Exporter, error) {
	switch cfg.Metrics.ExportMethod {
	case "":
		return nil, nil
	case "HTTP":
		return otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpoint(cfg.Metrics.Endpoint),
			otlpmetrichttp.WithTimeout(common.MillisToDuration(cfg.Metrics.Timeout)),
		)
	case "GRPC":
		return otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(cfg.Metrics.Endpoint),
			otlpmetricgrpc.WithTimeout(common.MillisToDuration(cfg.Metrics.Timeout)),
			otlpmetricgrpc.WithReconnectionPeriod(common.MillisToDuration(cfg.Metrics.Timeout)),
		)
	default:
		return nil, fmt.Errorf("cannot configure OpenTelemetry metric exporter for method=%s", cfg.Metrics.ExportMethod)
	}
}

func getTracerProviderFrom(ctx context.Context, cfg config.OpenTelemetry) *trace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		zap.L().Error("Cannot configure OpenTelemetry trace provider",
			zap.String("method", cfg.Spans.ExportMethod),
			zap.String("service", cfg.ServiceName),
			zap.String("endpoint", cfg.Spans.Endpoint),
			zap.Error(err),
		)
		return nil
	}
	exp, err := getTracerExporterFrom(ctx, cfg)
	if err != nil {
		zap.L().Error("Cannot configure OpenTelemetry trace provider",
			zap.String("method", cfg.Spans.ExportMethod),
			zap.String("service", cfg.ServiceName),
			zap.String("endpoint", cfg.Spans.Endpoint),
			zap.Error(err),
		)
		return nil
	}
	if exp == nil {
		zap.L().Warn("Cannot configure OpenTelemetry trace provider, method not specified")
		return nil
	}
	maxQueueSize := trace.DefaultMaxQueueSize
	if cfg.Spans.MaxQueueSize > -1 {
		maxQueueSize = cfg.Spans.MaxQueueSize
	}
	return trace.NewTracerProvider(
		trace.WithResource(r),
		trace.WithBatcher(exp,
			trace.WithMaxQueueSize(maxQueueSize),
		),
	)
}

func getTracerExporterFrom(ctx context.Context, cfg config.OpenTelemetry) (trace.SpanExporter, error) {
	switch cfg.Spans.ExportMethod {
	case "":
		return nil, nil
	case "HTTP":
		return otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(cfg.Spans.Endpoint),
			otlptracehttp.WithTimeout(common.MillisToDuration(cfg.Spans.Timeout)),
		)
	case "GRPC":
		return otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.Spans.Endpoint),
			otlptracegrpc.WithTimeout(common.MillisToDuration(cfg.Spans.Timeout)),
			otlptracegrpc.WithReconnectionPeriod(common.MillisToDuration(cfg.Spans.Timeout)),
		)
	default:
		return nil, fmt.Errorf("cannot configure OpenTelemetry span exporter for method=%s", cfg.Spans.ExportMethod)
	}
}
