package main

import (
	"log"
	"net"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/build"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/document"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation"
	"google.golang.org/grpc"
)

func main() {
	logfile := operation.UseLogFileInGenerated()
	defer operation.CloseLogFile(logfile)

	log.Println("Version " + build.Version)

	StartGrpcServer()
}

func StartGrpcServer() {
	listener, err := net.Listen("tcp", ":3110")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	RegisterAllGrpcServices(grpcServer)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

func RegisterAllGrpcServices(grpcServer *grpc.Server) {
	documentServiceApi.RegisterDocumentServiceServer(grpcServer, &document.DocumentServiceServer{})
}
