package metric

import (
	"testing"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
)

func InitializeMetrics(t *testing.T) {
	t.Setenv(metric.ServiceNameEnv, "test")
	t.Setenv(metric.OtelMetricEndpointEnv, "http://localhost:4317")
	metric.InitializeMetrics()
}
