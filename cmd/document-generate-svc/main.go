package main

import (
	"log"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/build"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/document"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	logPanic recovery.RecoveryHandlerFunc
)

func main() {
	log.Println("Version " + build.Version)

	StartGrpcServer()
}

func StartGrpcServer() {
	listener, err := net.Listen("tcp", ":3110")
	if err != nil {
		log.Fatal(err)
	}

	// Handling of panic to prevent crash from example nil pointer exceptions
	logPanic = func(p any) (err error) {
		log.Println(p)
		return status.Errorf(codes.Internal, "Internal server error occured.")
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(logPanic),
	}

	// router.RunListener(listener)
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
		log.Fatal(err)
	}
}

func RegisterAllGrpcServices(grpcServer *grpc.Server) {
	documentServiceApi.RegisterDocumentServiceServer(grpcServer, &document.DocumentServiceServer{})
}
