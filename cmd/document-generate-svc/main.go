package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
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

func main() {
	logger.SetLogLevel(zerolog.WarnLevel)
	logger.Logger.Info().Msg("Starting application.")

	provider, err := metric.InitializeMetrics()
	if err != nil {
		logger.Logger.Error().Err(err).Msg("failed to initialize metrics")
		os.Exit(40)
	}
	grpcServerStop := make(chan struct{})
	go startGrpcServer(grpcServerStop, "3110")

	<-grpcServerStop
	provider.Shutdown(context.Background())
	logger.Logger.Info().Msg("Application stopped.")
	os.Exit(0)
}

func startGrpcServer(grpcServerStop chan struct{}, port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Logger.Error().Err(err).Msgf("Failed to listen on port %v", port)
		os.Exit(50)
	}

	// Handling of panic to prevent crash from example nil pointer exceptions
	grpcServer := configureGrpcServer()

	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
	logger.Logger.Debug().Msg("starting grpc server")

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Logger.Error().Err(err).Msg("failed to start grpc server")
			os.Exit(51)
		}
	}()

	stop := <-gracefulStop
	grpcServer.GracefulStop()

	logger.Logger.Debug().Msgf("http server stopped. Received signal %s", stop)
	close(grpcServerStop)
}

func configureGrpcServer() *grpc.Server {
	// Handling of panic to prevent crash from example nil pointer exceptions
	logPanic := func(p any) (err error) {
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
	return grpcServer
}

func RegisterAllGrpcServices(grpcServer *grpc.Server) {
	documentServiceApi.RegisterDocumentServiceServer(grpcServer, &document.DocumentServiceServer{})
}
