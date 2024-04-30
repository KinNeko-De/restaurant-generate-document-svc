package metric

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/kinneko-de/restaurant-document-generate-svc/build"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/go-logr/zerologr"
)

const ServiceNameEnv = "OTEL_SERVICE_NAME"
const OtelMetricEndpointEnv = "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"

const MetricNameDocumentPreviewRequested = "restaurant.documents.preview.requested"
const MetricDescriptionDocumentPreviewRequested = "Sum of requested document previews"
const MetricNameDocumentPreviewDelivered = "restaurant.documents.preview.delivered"
const MetricDescriptionDocumentPreviewDelivered = "Sum of document previews that was delivered fully to the client"
const MetricNameDocumentGenerateSuccessful = "restaurant.documents.generate.successful"
const MetricDescriptionDocumentGenerateSuccessful = "Sum of documents that were generated successfully"
const MetricNameDocumentGenerateFailed = "restaurant.documents.generate.failed"
const MetricDescriptionDocumentGenerateFailed = "Sum of documents that failed to generate due to an error"
const MetricNameDocumentGenerateDuration = "restaurant.documents.generate.duration" // "Duration of document generation" Unit: "ms" Histogram
const MetricDescriptionDocumentGenerateDuration = "The duration of the document generation"
const MetricAttributeDocumentType = "document_type"

var (
	config                     otelConfig
	version                    = "0.2.0"
	ctx                        = context.Background()
	provider                   *metric.MeterProvider
	meter                      api.Meter
	previewRequested           api.Int64Counter
	previewDelivered           api.Int64Counter
	documentGenerateSuccessful api.Int64Counter
	documentGenerateFailed     api.Int64Counter
	documentGenerateDuration   api.Float64Histogram
)

func InitializeMetrics() (*metric.MeterProvider, error) {
	metricLogger := zerologr.New(&logger.Logger)
	otel.SetLogger(metricLogger)

	err := readConfig()
	if err != nil {
		return nil, err
	}

	provider, err := initializeOpenTelemetry()
	return provider, err
}

func initializeOpenTelemetry() (*metric.MeterProvider, error) {
	ressource, err := createRessource()
	if err != nil {
		return nil, err
	}

	readers, err := createReader()
	if err != nil {
		return nil, err
	}

	views := createViews()
	provider := createProvider(ressource, readers, views)
	metricError := createMetrics(provider)
	return provider, metricError
}

func createRessource() (*resource.Resource, error) {
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.OtelServiceName),
			semconv.ServiceVersionKey.String(build.Version),
		))
	if err != nil {
		return nil, fmt.Errorf("failed to create ressource for metric reader: %w", err)
	}

	return res, nil
}

func createViews() []metric.View {
	view := metric.NewView(
		metric.Instrument{
			Name: MetricNameDocumentGenerateDuration,
			Kind: metric.InstrumentKindHistogram,
		},
		metric.Stream{
			Aggregation: metric.AggregationExplicitBucketHistogram{
				NoMinMax:   true,
				Boundaries: []float64{1000, 4000, 7000, 10000, 20000},
			},
		},
	)

	return []metric.View{view}
}

func createReader() ([]metric.Reader, error) {
	otelGrpcExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure(), otlpmetricgrpc.WithEndpoint(config.OtelMetricEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metric reader to otel collector: %w", err)
	}
	otelReader := metric.NewPeriodicReader(otelGrpcExporter)

	consoleExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metric reader to console: %w", err)
	}
	consoleReader := metric.NewPeriodicReader(consoleExporter)

	return []metric.Reader{otelReader, consoleReader}, nil
}

func createProvider(ressource *resource.Resource, readers []metric.Reader, views []metric.View) *metric.MeterProvider {
	options := []metric.Option{
		metric.WithResource(ressource),
		metric.WithView(views...),
	}
	for _, reader := range readers {
		options = append(options, metric.WithReader(reader))
	}

	provider = metric.NewMeterProvider(
		options...,
	)
	otel.SetMeterProvider(provider)
	return provider
}

// https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/
func createMetrics(provider *metric.MeterProvider) error {
	// I decided to use the service name here as scope because this service is a microservice. one sccope per service approach.
	meter = provider.Meter(config.OtelServiceName, api.WithInstrumentationVersion(version))

	var err error
	errorTemplate := "failed to initialize metric '%v' %w"
	previewRequested, err = meter.Int64Counter(
		MetricNameDocumentPreviewRequested,
		api.WithDescription(MetricDescriptionDocumentPreviewRequested))
	if err != nil {
		return fmt.Errorf(errorTemplate, MetricNameDocumentPreviewRequested, err)
	}
	previewDelivered, err = meter.Int64Counter(
		MetricNameDocumentPreviewDelivered,
		api.WithDescription(MetricDescriptionDocumentPreviewDelivered))
	if err != nil {
		return fmt.Errorf(errorTemplate, MetricNameDocumentPreviewDelivered, err)
	}

	documentGenerateSuccessful, err = meter.Int64Counter(
		MetricNameDocumentGenerateSuccessful,
		api.WithDescription(MetricDescriptionDocumentGenerateSuccessful))
	if err != nil {
		return fmt.Errorf(errorTemplate, MetricNameDocumentGenerateSuccessful, err)
	}
	documentGenerateFailed, err = meter.Int64Counter(
		MetricNameDocumentGenerateFailed,
		api.WithDescription(MetricDescriptionDocumentGenerateFailed))
	if err != nil {
		return fmt.Errorf(errorTemplate, MetricNameDocumentGenerateFailed, err)
	}

	documentGenerateDuration, err = meter.Float64Histogram(
		MetricNameDocumentGenerateDuration,
		api.WithDescription(MetricDescriptionDocumentGenerateDuration),
		api.WithUnit("ms"))
	if err != nil {
		return fmt.Errorf(errorTemplate, MetricNameDocumentGenerateDuration, err)
	}

	return nil
}

func PreviewRequested() {
	previewRequested.Add(ctx, 1)
}

func PreviewDelivered() {
	previewDelivered.Add(ctx, 1)
}

func DocumentGenerated(documentType string, duration time.Duration, err error) {
	if err != nil {
		documentGenerateFailed.Add(ctx, 1, api.WithAttributes(attribute.Key(MetricAttributeDocumentType).String(documentType)))
	} else {
		documentGenerateSuccessful.Add(ctx, 1, api.WithAttributes(attribute.Key(MetricAttributeDocumentType).String(documentType)))
		documentGenerateDuration.Record(ctx, float64(duration.Milliseconds()), api.WithAttributes(attribute.Key(MetricAttributeDocumentType).String(documentType)))
	}
}

func ForceFlush() {
	provider.ForceFlush(ctx)
}

func readConfig() error {
	otelConfig, err := loadConfig()
	if err != nil {
		return err
	}
	config = otelConfig

	return nil
}

type otelConfig struct {
	OtelMetricEndpoint string // is used by the otel sdk to identify the endpoint to send metrics to. According to document it Will be set implicitly by the otel sdk. But it does not work. I set it explicitly.
	OtelServiceName    string // is used by the otel sdk to identify the service name. I found no way to set it explicitly by the otel sdk. According to the specification setting an attribute with name "service.name" should work, but it does not.
}

func loadConfig() (otelConfig, error) {
	endpoint, found := os.LookupEnv(OtelMetricEndpointEnv)
	if !found {
		return otelConfig{}, fmt.Errorf("otel metric endpoint is not configured. Expected environment variable %v", OtelMetricEndpointEnv)
	}

	serviceName, found := os.LookupEnv(ServiceNameEnv)
	if !found {
		return otelConfig{}, fmt.Errorf("otel service name is not configured. Expected environment variable %v", ServiceNameEnv)
	}

	return otelConfig{
		OtelMetricEndpoint: endpoint,
		OtelServiceName:    serviceName,
	}, nil
}
