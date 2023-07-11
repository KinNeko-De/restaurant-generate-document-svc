package main

import (
	"log"
	"net"

	"github.com/KinNeko-De/restaurant-document-svc/internal/app/document"
	"github.com/KinNeko-De/restaurant-document-svc/internal/app/operation"
	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"google.golang.org/grpc"
)

func main() {
	logfile := operation.UseLogFileInGenerated()
	defer operation.CloseLogFile(logfile)

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
