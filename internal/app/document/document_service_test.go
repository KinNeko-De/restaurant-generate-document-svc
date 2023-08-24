package document

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	iomocks "github.com/kinneko-de/restaurant-document-generate-svc/internal/testing/io/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGeneratePreview_InvalidRequests(t *testing.T) {
	ctx := context.Background()
	client, closer := server(ctx)
	defer closer()

	type expectation struct {
		error
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
					status.Error(codes.InvalidArgument, "requested document is mandatory to generate a document."),
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
				assert.Nil(t, actualResponse)
				assert.Equal(t, expected.error, actualError)
			}
		})
	}
}

func TestGeneratePreview_DocumentIsGenerated(t *testing.T) {
	expectedFileSize := uint64(13354)
	expectedMediaType := "application/pdf"
	expectedExtension := ".pdf"

	mockReader := iomocks.NewReader(t)
	mockReader.EXPECT().Read(mock.Anything).Return(100, nil).Once()
	mockReader.EXPECT().Read(mock.Anything).Return(100, nil).Once()
	mockReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	mockGenerator := NewDocumentGeneratorMock(t)
	mockFileHandler := NewFileHandlerMock(t)
	generatedFile := GeneratedFile{
		Size:    int64(expectedFileSize),
		Reader:  bufio.NewReader(mockReader),
		Handler: mockFileHandler,
	}
	mockGenerator.EXPECT().GenerateDocument(mock.Anything, mock.Anything).Return(generatedFile, nil)
	documentGenerator = mockGenerator
	mockFileHandler.EXPECT().Close().Return(nil)
	ctx := context.Background()
	client, closer := server(ctx)
	defer closer()

	request := &documentServiceApi.GeneratePreviewRequest{
		RequestedDocument: &documentServiceApi.RequestedDocument{
			Type: &documentServiceApi.RequestedDocument_Invoice{},
		},
	}

	expected := []*documentServiceApi.GeneratePreviewResponse{
		{
			File: &documentServiceApi.GeneratePreviewResponse_Metadata{
				Metadata: &documentServiceApi.GeneratedFileMetadata{
					Size:      expectedFileSize,
					MediaType: expectedMediaType,
					Extension: expectedExtension,
				},
			},
		},
		{
			File: &documentServiceApi.GeneratePreviewResponse_Chunk{
				Chunk: make([]byte, 100),
			},
		},
		{
			File: &documentServiceApi.GeneratePreviewResponse_Chunk{
				Chunk: make([]byte, 100),
			},
		},
	}

	stream, err := client.GeneratePreview(ctx, request)
	assert.NotNil(t, stream)
	assert.Nil(t, err)

	actualFirstResponse, actualError := stream.Recv()
	assert.Equal(t, nil, actualError)
	assert.NotNil(t, actualFirstResponse)
	actualMetadataResponse := actualFirstResponse.GetMetadata()
	expectedMetadataResponse := expected[0].GetMetadata()
	// TODO replace with require
	assert.NotNil(t, actualMetadataResponse)
	assert.NotNil(t, actualMetadataResponse.CreatedAt)
	assert.Equal(t, actualMetadataResponse.MediaType, expectedMetadataResponse.MediaType)
	assert.Equal(t, actualMetadataResponse.Extension, expectedMetadataResponse.Extension)
	assert.Equal(t, actualMetadataResponse.Size, expectedMetadataResponse.Size)

	for _, expectedResponse := range expected[1:] {
		actualResponse, actualError := stream.Recv()
		assert.Equal(t, nil, actualError)
		assert.NotNil(t, actualResponse)
		actualChunk := actualResponse.GetChunk()
		assert.Equal(t, expectedResponse.GetChunk(), actualChunk)
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
