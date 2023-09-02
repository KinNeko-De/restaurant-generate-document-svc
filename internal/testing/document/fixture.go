package document

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation"
)

func CreateDocumentServiceClient(ctx context.Context, server documentServiceApi.DocumentServiceServer) (documentServiceApi.DocumentServiceClient, func()) {
	logger := operation.Logger.With().Ctx(ctx).Logger()
	buffer := 65536 // 64 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	documentServiceApi.RegisterDocumentServiceServer(baseServer, server)
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			logger.Error().Msgf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error().Msgf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			logger.Error().Msgf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := documentServiceApi.NewDocumentServiceClient(conn)

	return client, closer
}
