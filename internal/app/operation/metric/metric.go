package metric

import (
	"context"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

var (
	ctx               = context.Background()
	reader            metric.Exporter
	provider          *metric.MeterProvider
	meter             api.Meter
	documentRequested api.Int64Counter
	documentGenerated api.Int64Counter
	documentFailed    api.Int64Counter
)

func InitializeMetrics() (err error) {
	reader, err = stdoutmetric.New()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize metric reader")
	}
	provider = metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(reader)))
	meter = provider.Meter("restaurant-document-generate-svc", api.WithInstrumentationVersion("0.1.0"))

	documentRequested, err = meter.Int64Counter("document-requested", api.WithUnit("document"), api.WithDescription("Number of documents requested"))
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize metric 'document-requested'")
	}

	documentGenerated, err = meter.Int64Counter("document-generated", api.WithUnit("document"), api.WithDescription("Number of documents successfully generated"))
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize metric 'document-generated'")
	}

	documentFailed, err = meter.Int64Counter("document-failed", api.WithUnit("document"), api.WithDescription("Number of documents that can not generated because of error"))
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize metric 'document-failed'")
	}
	return
}

func DocumentRequested(documentType string) {
	documentRequested.Add(ctx, 1, api.WithAttributes(attribute.Key("document_type").String(documentType)))
}

func DocumentGenerated(documentType string) {
	documentGenerated.Add(ctx, 1, api.WithAttributes(attribute.Key("document_type").String(documentType)))
}

func DocumentFailed(documentType string) {
	documentFailed.Add(ctx, 1, api.WithAttributes(attribute.Key("document_type").String(documentType)))
}

func ForceFlush() {
	provider.ForceFlush(ctx)
}
