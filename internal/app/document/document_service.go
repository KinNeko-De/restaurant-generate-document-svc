package document

import (
	"io"
	"log"
	"time"

	"github.com/google/uuid"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const chunkSize = 43008 // 42 * 1024

type DocumentServiceServer struct {
	documentServiceApi.UnimplementedDocumentServiceServer
}

func (s *DocumentServiceServer) GeneratePreview(request *documentServiceApi.GeneratePreviewRequest, stream documentServiceApi.DocumentService_GeneratePreviewServer) error {
	start := time.Now()
	requestId := uuid.New()

	if request.RequestedDocument == nil {
		return status.Error(codes.InvalidArgument, "requested document is mandatory to generate a document.")
	}

	log.Println("Preprocessing: " + time.Since(start).String())
	start = time.Now()

	result, err := GenerateDocument(requestId, request.RequestedDocument)
	if err != nil {
		log.Println(err) // TODO make this debug
		return status.Error(codes.Internal, "generation of document failed.")
	}
	log.Println("Generation: " + time.Since(start).String())
	start = time.Now()

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

		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Error(codes.Internal, "generation of document failed.")
		}
	}

	log.Println("Sneding: " + time.Since(start).String())

	return nil
}

/*
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
*/
