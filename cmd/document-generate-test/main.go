package main

import (
	"context"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/testing"
	"github.com/rs/zerolog"
)

func main() {
	logger.SetLogLevel(zerolog.DebugLevel)
	provider, err := metric.InitializeMetrics()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("failed to initialize metrics")
	}
	defer provider.Shutdown(context.Background())
	testing.GenerateTestInvoice()
	metric.ForceFlush() // otherwise, the metrics will not be sent to the collector and console
}
