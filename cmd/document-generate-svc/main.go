package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/health"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
)

func main() {
	logger.SetLogLevel(zerolog.DebugLevel)
	logger.Logger.Info().Msg("Starting application.")

	provider, err := metric.InitializeMetrics()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to initialize metrics")
		os.Exit(40)
	}
	grpcServerStop := make(chan struct{})
	grpcServerStarted := make(chan struct{})
	go startGrpcServer(grpcServerStop, grpcServerStarted, "3110")

	<-grpcServerStarted
	health.Ready()

	<-grpcServerStop
	provider.Shutdown(context.Background())
	logger.Logger.Info().Msg("Application stopped.")
	os.Exit(0)
}
