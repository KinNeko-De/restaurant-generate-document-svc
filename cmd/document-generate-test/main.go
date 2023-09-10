package main

import (
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/testing"
	"github.com/rs/zerolog"
)

func main() {
	logger.SetLogLevel(zerolog.DebugLevel)
	metric.InitializeMetrics()
	testing.GenerateTestInvoice()
	metric.ForceFlush() // otherwise, the metrics will not be sent to the collector and console
}
