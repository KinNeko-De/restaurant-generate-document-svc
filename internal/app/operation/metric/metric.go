package metric

import (
	"context"
	"fmt"
	"os"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/kinneko-de/restaurant-document-generate-svc/build"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/go-logr/zerologr"
)

const ServiceNameEnv = "OTEL_SERVICE_NAME"
const OtelMetricEndpointEnv = "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"

var (
	config            otelConfig
	version           = "0.1.0"
	ctx               = context.Background()
	provider          *metric.MeterProvider
	meter             api.Meter
	documentRequested api.Int64Counter
	documentGenerated api.Int64Counter
	documentFailed    api.Int64Counter
	documentDelivered api.Int64Counter
)

func InitializeMetrics() error {
	metricLogger := zerologr.New(&logger.Logger)
	otel.SetLogger(metricLogger)

	err := readConfig()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to read metric reader configuration")
	}

	ressource := createRessource()
	provider := createProvider(ressource)
	createMetrics(provider)

	return nil
}

func createRessource() *resource.Resource {
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.OtelServiceName),
			semconv.ServiceVersionKey.String(build.Version),
		))
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to create ressource for metric reader")
	}

	return res
}

func createProvider(ressource *resource.Resource) *metric.MeterProvider {
	otelReader, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure(), otlpmetricgrpc.WithEndpoint(config.OtelMetricEndpoint))
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize metric reader to otel collector")
	}
	consoleReader, err := stdoutmetric.New()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize metric reader to console")
	}

	provider = metric.NewMeterProvider(
		metric.WithResource(ressource),
		metric.WithReader(metric.NewPeriodicReader(consoleReader)),
		metric.WithReader(metric.NewPeriodicReader(otelReader)),
	)
	otel.SetMeterProvider(provider)
	return provider
}

// TODO make name more according to name in https://github.com/MrAlias/opentelemetry-go-contrib/blob/main/instrumentation/net/http/otelhttp/test/handler_test.go "		Name: "http.server.request_content_length" and https://opentelemetry.io/docs/specs/otel/metrics/semantic_conventions/
func createMetrics(provider *metric.MeterProvider) {
	// I decided to use the service name here as scope because this service is a microservice. one sccope per service approach.
	meter = provider.Meter(config.OtelServiceName, api.WithInstrumentationVersion(version))

	var err error

	documentRequested, err = meter.Int64Counter("document-requested", api.WithUnit("document"), api.WithDescription("Number of requested documents"))
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

	documentDelivered, err = meter.Int64Counter("document-delivered", api.WithUnit("document"), api.WithDescription("Number of documents that was delivered fully to the client"))
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("Failed to initialize metric 'document-delivered'")
	}
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

func DocumentDelivered(documentType string) {
	documentDelivered.Add(ctx, 1, api.WithAttributes(attribute.Key("document_type").String(documentType)))
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
