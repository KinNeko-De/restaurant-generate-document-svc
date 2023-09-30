package document

import (
	"bufio"
	"context"
	"errors"
	"io"
	"testing"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	documentfixture "github.com/kinneko-de/restaurant-document-generate-svc/internal/testing/document"
	documentmocks "github.com/kinneko-de/restaurant-document-generate-svc/internal/testing/document/mocks"

	iomocks "github.com/kinneko-de/restaurant-document-generate-svc/internal/testing/io/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGeneratePreview_DocumentIsGenerated(t *testing.T) {
	expectedFileSize := uint64(chunkSize + 100)
	expectedMediaType := "application/pdf"
	expectedExtension := ".pdf"

	mockReader := iomocks.NewReader(t)
	mockGenerator := NewMockDocumentGenerator(t)
	mockFileHandler := NewMockFileHandler(t)
	generatedFile := GeneratedFile{
		Size:    int64(expectedFileSize),
		Reader:  bufio.NewReader(mockReader),
		Handler: mockFileHandler,
	}
	mockGenerator.EXPECT().GenerateDocument(mock.Anything, mock.Anything, mock.Anything).Return(generatedFile, nil)
	mockReader.EXPECT().Read(mock.Anything).Return(chunkSize, nil).Once()
	mockReader.EXPECT().Read(mock.Anything).Return(100, nil).Once()
	mockReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	mockFileHandler.EXPECT().Close().Return(nil).Once()
	documentGenerator = mockGenerator
	ctx := context.Background()
	client, closer := documentfixture.CreateDocumentServiceClient(ctx, &DocumentServiceServer{})
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
				Chunk: make([]byte, chunkSize),
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
	require.Nil(t, actualError)
	require.NotNil(t, actualFirstResponse)
	actualMetadataResponse := actualFirstResponse.GetMetadata()
	expectedMetadataResponse := expected[0].GetMetadata()
	require.NotNil(t, actualMetadataResponse)
	assert.NotNil(t, actualMetadataResponse.CreatedAt)
	assert.Equal(t, actualMetadataResponse.MediaType, expectedMetadataResponse.MediaType)
	assert.Equal(t, actualMetadataResponse.Extension, expectedMetadataResponse.Extension)
	assert.Equal(t, actualMetadataResponse.Size, expectedMetadataResponse.Size)
	for _, expectedResponse := range expected[1:] {
		actualResponse, actualError := stream.Recv()
		require.Nil(t, actualError)
		assert.NotNil(t, actualResponse)
		actualChunk := actualResponse.GetChunk()
		assert.Equal(t, expectedResponse.GetChunk(), actualChunk)
	}
	_, endOfDtreamError := stream.Recv()
	assert.Equal(t, io.EOF, endOfDtreamError)
	closeErr := stream.CloseSend()
	assert.Nil(t, closeErr)
}

func TestGeneratePreview_InvalidRequests(t *testing.T) {
	ctx := context.Background()
	client, closer := documentfixture.CreateDocumentServiceClient(ctx, &DocumentServiceServer{})
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

func TestGeneratePreview_GenerateDocumentFailed(t *testing.T) {
	expected := codes.Internal

	mockStream := documentmocks.NewDocumentService_GeneratePreviewServer(t)
	request := &documentServiceApi.GeneratePreviewRequest{
		RequestedDocument: &documentServiceApi.RequestedDocument{
			Type: &documentServiceApi.RequestedDocument_Invoice{},
		},
	}
	mockGenerator := NewMockDocumentGenerator(t)
	mockGenerator.EXPECT().GenerateDocument(mock.Anything, mock.Anything, mock.Anything).Return(GeneratedFile{}, errors.New("TestError"))
	documentGenerator = mockGenerator

	server := DocumentServiceServer{}
	actualError := server.GeneratePreview(request, mockStream)
	require.NotNil(t, actualError)
	actual := status.Code(actualError)
	assert.Equal(t, expected, actual)
}

func TestGeneratePreview_SendMetadataFailed(t *testing.T) {
	expected := codes.Internal

	mockStream := documentmocks.NewDocumentService_GeneratePreviewServer(t)
	request := &documentServiceApi.GeneratePreviewRequest{
		RequestedDocument: &documentServiceApi.RequestedDocument{
			Type: &documentServiceApi.RequestedDocument_Invoice{},
		},
	}
	mockReader := iomocks.NewReader(t)
	mockGenerator := NewMockDocumentGenerator(t)
	mockFileHandler := NewMockFileHandler(t)
	generatedFile := GeneratedFile{
		Size:    int64(544),
		Reader:  bufio.NewReader(mockReader),
		Handler: mockFileHandler,
	}
	mockGenerator.EXPECT().GenerateDocument(mock.Anything, mock.Anything, mock.Anything).Return(generatedFile, nil)
	mockStream.EXPECT().Send(mock.Anything).Return(errors.New("Network error")).Once()
	mockFileHandler.EXPECT().Close().Return(nil).Once()
	documentGenerator = mockGenerator

	server := DocumentServiceServer{}
	actualError := server.GeneratePreview(request, mockStream)
	require.NotNil(t, actualError)
	actual := status.Code(actualError)
	assert.Equal(t, expected, actual)
}

func TestGeneratePreview_SendChunkFailed(t *testing.T) {
	expected := codes.Internal

	mockStream := documentmocks.NewDocumentService_GeneratePreviewServer(t)
	request := &documentServiceApi.GeneratePreviewRequest{
		RequestedDocument: &documentServiceApi.RequestedDocument{
			Type: &documentServiceApi.RequestedDocument_Invoice{},
		},
	}
	mockReader := iomocks.NewReader(t)
	mockGenerator := NewMockDocumentGenerator(t)
	mockFileHandler := NewMockFileHandler(t)
	generatedFile := GeneratedFile{
		Size:    int64(544),
		Reader:  bufio.NewReader(mockReader),
		Handler: mockFileHandler,
	}
	mockGenerator.EXPECT().GenerateDocument(mock.Anything, mock.Anything, mock.Anything).Return(generatedFile, nil)
	mockReader.EXPECT().Read(mock.Anything).Return(1, nil).Once()
	mockStream.EXPECT().Send(mock.Anything).Return(nil).Once()
	mockStream.EXPECT().Send(mock.Anything).Return(errors.New("Network error")).Once()
	mockFileHandler.EXPECT().Close().Return(nil).Once()
	documentGenerator = mockGenerator

	server := DocumentServiceServer{}
	actualError := server.GeneratePreview(request, mockStream)
	require.NotNil(t, actualError)
	actual := status.Code(actualError)
	assert.Equal(t, expected, actual)
}

func TestGeneratePreview_ReadingFileFailed(t *testing.T) {
	expected := codes.Internal

	mockStream := documentmocks.NewDocumentService_GeneratePreviewServer(t)
	request := &documentServiceApi.GeneratePreviewRequest{
		RequestedDocument: &documentServiceApi.RequestedDocument{
			Type: &documentServiceApi.RequestedDocument_Invoice{},
		},
	}
	mockReader := iomocks.NewReader(t)
	mockGenerator := NewMockDocumentGenerator(t)
	mockFileHandler := NewMockFileHandler(t)
	generatedFile := GeneratedFile{
		Size:    int64(544),
		Reader:  bufio.NewReader(mockReader),
		Handler: mockFileHandler,
	}
	mockGenerator.EXPECT().GenerateDocument(mock.Anything, mock.Anything, mock.Anything).Return(generatedFile, nil)
	mockStream.EXPECT().Send(mock.Anything).Return(nil).Once()
	mockReader.EXPECT().Read(mock.Anything).Return(0, errors.New("Reading file failed")).Once()
	mockFileHandler.EXPECT().Close().Return(nil).Once()
	documentGenerator = mockGenerator

	server := DocumentServiceServer{}
	actualError := server.GeneratePreview(request, mockStream)
	require.NotNil(t, actualError)
	actual := status.Code(actualError)
	assert.Equal(t, expected, actual)
}

func TestGeneratePreview_CLosingFileFailed_ErrorIsIgnored(t *testing.T) {
	mockStream := documentmocks.NewDocumentService_GeneratePreviewServer(t)
	request := &documentServiceApi.GeneratePreviewRequest{
		RequestedDocument: &documentServiceApi.RequestedDocument{
			Type: &documentServiceApi.RequestedDocument_Invoice{},
		},
	}
	mockReader := iomocks.NewReader(t)
	mockGenerator := NewMockDocumentGenerator(t)
	mockFileHandler := NewMockFileHandler(t)
	generatedFile := GeneratedFile{
		Size:    int64(544),
		Reader:  bufio.NewReader(mockReader),
		Handler: mockFileHandler,
	}
	mockGenerator.EXPECT().GenerateDocument(mock.Anything, mock.Anything, mock.Anything).Return(generatedFile, nil)
	mockStream.EXPECT().Send(mock.Anything).Return(nil).Once()
	mockReader.EXPECT().Read(mock.Anything).Return(0, io.EOF).Once()
	mockFileHandler.EXPECT().Close().Return(errors.New("Closing file failed.")).Once()
	documentGenerator = mockGenerator

	server := DocumentServiceServer{}
	actualError := server.GeneratePreview(request, mockStream)
	assert.Nil(t, actualError)
}

// / This test is not really necessary, but it is here to show that the code is not broken
func TestGeneratePreview_ReadReturnsZeroBytesButNoError(t *testing.T) {
	expected := codes.Internal

	mockStream := documentmocks.NewDocumentService_GeneratePreviewServer(t)
	expectedFileSize := uint64(chunkSize + 100)
	mockReader := iomocks.NewReader(t)
	mockGenerator := NewMockDocumentGenerator(t)
	mockFileHandler := NewMockFileHandler(t)
	generatedFile := GeneratedFile{
		Size:    int64(expectedFileSize),
		Reader:  bufio.NewReader(mockReader),
		Handler: mockFileHandler,
	}
	mockGenerator.EXPECT().GenerateDocument(mock.Anything, mock.Anything, mock.Anything).Return(generatedFile, nil)
	mockStream.EXPECT().Send(mock.Anything).Return(nil).Once()
	mockReader.EXPECT().Read(mock.Anything).Return(0, nil).Once()
	mockFileHandler.EXPECT().Close().Return(nil).Once()
	documentGenerator = mockGenerator

	request := &documentServiceApi.GeneratePreviewRequest{
		RequestedDocument: &documentServiceApi.RequestedDocument{
			Type: &documentServiceApi.RequestedDocument_Invoice{},
		},
	}

	server := DocumentServiceServer{}
	actualError := server.GeneratePreview(request, mockStream)

	assert.NotNil(t, actualError)
	require.NotNil(t, actualError)
	actual := status.Code(actualError)
	assert.Equal(t, expected, actual)
}
