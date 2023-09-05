package operation

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

var (
	ctx               = context.Background()
	provider          = metric.NewMeterProvider()
	meter             = provider.Meter("restaurant-document-generate-svc", api.WithInstrumentationVersion("0.1.0"))
	documentRequested api.Int64Counter
)

func InitializeMetrics() {
	newDocumentRequested, err := meter.Int64Counter("document-requested", api.WithUnit("document"), api.WithDescription("Number of documents requested"))
	documentRequested = newDocumentRequested
	if err != nil {
		Logger.Fatal().Err(err).Msg("Failed to initialize metric 'document-requested'")
	}
}

func DocumentRequested(documentType string) {
	documentRequested.Add(ctx, 1, api.WithAttributes(attribute.Key("document_type").String(documentType)))
}
