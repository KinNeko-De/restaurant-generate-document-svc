package main

import (
	"context"
	"net"

	"github.com/rs/zerolog/log"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/document"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	logPanic recovery.RecoveryHandlerFunc
)

func main() {
	logger.SetInfoLogLevel()

	provider, err := metric.InitializeMetrics()
	if err != nil {
		logger.Logger.Fatal().Err(err).Msg("failed to initialize metrics")
	}
	defer provider.Shutdown(context.Background())

	StartGrpcServer()
}

func StartGrpcServer() *grpc.Server {
	port := "3110"
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Logger.Fatal().Err(err).Msgf("Failed to listen on port %v", port)
	}

	// Handling of panic to prevent crash from example nil pointer exceptions
	logPanic = func(p any) (err error) {
		log.Error().Any("method", p).Err(err).Msg("Recovered from panic.")
		return status.Errorf(codes.Internal, "Internal server error occured.")
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(logPanic),
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			recovery.UnaryServerInterceptor(opts...),
		),
		grpc.StreamInterceptor(
			recovery.StreamServerInterceptor(opts...),
		),
	)
	RegisterAllGrpcServices(grpcServer)
	if err := grpcServer.Serve(listener); err != nil {
		logger.Logger.Fatal().Err(err).Msg("grpc server was aborted. Graceful shutdown should be implemented.")
	}

	return grpcServer
}

func RegisterAllGrpcServices(grpcServer *grpc.Server) {
	documentServiceApi.RegisterDocumentServiceServer(grpcServer, &document.DocumentServiceServer{})
}
