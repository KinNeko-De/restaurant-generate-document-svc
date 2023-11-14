package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/health"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/server"
)

func main() {
	logger.SetLogLevel(zerolog.WarnLevel)
	logger.Logger.Info().Msg("Starting application.")

	provider, err := metric.InitializeMetrics()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to initialize metrics")
		os.Exit(40)
	}
	grpcServerStopped := make(chan struct{})
	grpcServerStarted := make(chan struct{})
	go server.StartGrpcServer(grpcServerStopped, grpcServerStarted, ":3110")

	<-grpcServerStarted
	health.Ready()

	<-grpcServerStopped
	provider.Shutdown(context.Background())
	logger.Logger.Info().Msg("Application stopped.")
	os.Exit(0)
}
