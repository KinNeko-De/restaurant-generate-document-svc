package document

import (
	"context"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGeneratePreview(t *testing.T) {
	ctx := context.Background()
	client, closer := server(ctx)
	defer closer()

	type expectation struct {
		response *documentServiceApi.GeneratePreviewResponse
		err      error
	}

	tests := map[string]struct {
		request  *documentServiceApi.GeneratePreviewRequest
		expected []expectation
	}{
		"RequestedDocumentIsNil": {
			request: &documentServiceApi.GeneratePreviewRequest{
				RequestedDocument: nil,
			},
			expected: []expectation{
				{
					response: nil,
					err:      status.Error(codes.InvalidArgument, "requested document is mandatory to generate a document."),
				},
			},
		},
	}

	for scenario, test := range tests {
		t.Run(scenario, func(t *testing.T) {

			stream, err := client.GeneratePreview(ctx, test.request)
			assert.NotNil(t, stream)
			assert.Nil(t, err)

			for _, expected := range test.expected {
				actualResponse, actualError := stream.Recv()
				assert.Equal(t, expected.response, actualResponse)
				assert.Equal(t, expected.err, actualError)
			}
		})
	}
}

func server(ctx context.Context) (documentServiceApi.DocumentServiceClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	documentServiceApi.RegisterDocumentServiceServer(baseServer, &DocumentServiceServer{})
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	client := documentServiceApi.NewDocumentServiceClient(conn)

	return client, closer
}
