package metric

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"

	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"
)

func TestInitializeMetrics_ConfigIsComplete(t *testing.T) {
	expectedServiceName := "expectedServiceName"
	expectedOtelMetricEndpoint := "otel-collector:4317"
	t.Setenv(ServiceNameEnv, expectedServiceName)
	t.Setenv(OtelMetricEndpointEnv, expectedOtelMetricEndpoint)

	InitializeMetrics()

	assert.Equal(t, expectedServiceName, config.OtelServiceName)
	assert.Equal(t, expectedOtelMetricEndpoint, config.OtelMetricEndpoint)
	assert.NotNil(t, provider)
	assert.NotNil(t, meter)
	assert.NotNil(t, documentRequested)
	assert.NotNil(t, documentGenerated)
	assert.NotNil(t, documentFailed)
	assert.NotNil(t, documentDelivered)

}

func TestDocumentRequested_Invoice(t *testing.T) {
	expectedValue := int64(1)
	expectedDocumentType := "Invoice"
	expectedAttriutes := attribute.NewSet(attribute.String("document_type", expectedDocumentType))
	expectedMetric := metricdata.Metrics{
		Name:        "document-requested",
		Description: "Number of requested documents",
		Unit:        "document",
		Data: metricdata.Sum[int64]{
			DataPoints:  []metricdata.DataPoint[int64]{{Attributes: expectedAttriutes, Value: expectedValue}},
			Temporality: metricdata.CumulativeTemporality,
			IsMonotonic: true,
		},
	}

	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(metric.WithResource(createRessource()), metric.WithReader(reader))
	createMetrics(meterProvider)

	DocumentRequested(expectedDocumentType)

	rm := metricdata.ResourceMetrics{}
	err := reader.Collect(context.Background(), &rm)
	require.NoError(t, err)
	require.Len(t, rm.ScopeMetrics, 1)
	require.Len(t, rm.ScopeMetrics[0].Metrics, 1)
	metricdatatest.AssertEqual(t, expectedMetric, rm.ScopeMetrics[0].Metrics[0], metricdatatest.IgnoreTimestamp())
}
