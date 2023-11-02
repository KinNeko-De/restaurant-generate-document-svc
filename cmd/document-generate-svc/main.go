package main

import (
	"context"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/rs/zerolog"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/server"
)

var (
	logPanic recovery.RecoveryHandlerFunc
)

func main() {
	logger.SetLogLevel(zerolog.DebugLevel)
	logger.Logger.Info().Msg("Starting application.")

	provider, err := metric.InitializeMetrics()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("failed to initialize metrics")
	}
	grpcServerStop := make(chan struct{})
	server.StartGrpcServer(grpcServerStop, "3110")

	<-grpcServerStop
	provider.Shutdown(context.Background())
	logger.Logger.Info().Msg("Application stopped.")
	os.Exit(0)
}
