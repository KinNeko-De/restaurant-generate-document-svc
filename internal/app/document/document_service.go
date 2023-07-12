package document

import (
	"io"

	"github.com/google/uuid"
	"github.com/kinneko-de/api-contract/golang/kinnekode/protobuf"
	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DocumentServiceServer struct {
	documentServiceApi.UnimplementedDocumentServiceServer
}

func (s *DocumentServiceServer) GeneratePreview(request *documentServiceApi.GeneratePreviewRequest, stream documentServiceApi.DocumentService_GeneratePreviewServer) error {
	requestId, err := ParseRequestId(request)
	if err != nil {
		return err
	}

	result, err := GenerateDocument(requestId, request.RequestedDocument)
	if err != nil {
		return status.Error(codes.Internal, "generation of document failed.")
	}

	if err := stream.Send(&documentServiceApi.GeneratePreviewResponse{
		File: &documentServiceApi.GeneratePreviewResponse_Metadata{
			Metadata: &documentServiceApi.GeneratedFileMetadata{
				CreatedAt: timestamppb.Now(),
				Size:      uint64(result.Size),
				MediaType: "application/pdf",
				Extension: ".pdf",
			},
		},
	}); err != nil {
		return err
	}

	const chunkSize = 1000
	chunks := make([]byte, 0, chunkSize)
	for {
		numberOfReadBytes, err := result.Reader.Read(chunks[:cap(chunks)])
		if numberOfReadBytes > 0 {
			chunks = chunks[:numberOfReadBytes]
			if err := stream.Send(&documentServiceApi.GeneratePreviewResponse{
				File: &documentServiceApi.GeneratePreviewResponse_Chunk{
					Chunk: chunks,
				},
			}); err != nil {
				return err
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, "generation of document failed.")
		}
	}

	return nil
}

func ParseRequestId(request *documentServiceApi.GeneratePreviewRequest) (uuid uuid.UUID, grpcError error) {
	if request.RequestId == nil {
		return uuid, status.Error(codes.InvalidArgument, "requestId is mandatory to generate a document.")
	}
	requestId, err := protobuf.ToUuid(request.RequestId)
	if err != nil {
		return uuid, status.Error(codes.InvalidArgument, "requestId '"+request.RequestId.Value+"' is not a valid uuid. expect uuid in the following format: 550e8400-e29b-11d4-a716-446655440000")
	}
	return requestId, nil
}
