package document

import (
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
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
	logger := logger.Logger.With().Str("requestId", requestId.String()).Logger()

	if request.RequestedDocument == nil {
		return status.Error(codes.InvalidArgument, "requested document is mandatory to generate a document.")
	}

	logger.Debug().Msgf("Preprocessing: %v", time.Since(start))
	start = time.Now()

	result, err := GenerateDocument(requestId, request.RequestedDocument, logger)
	if err != nil {
		return status.Error(codes.Internal, "Generation of document failed.")
	}
	defer CloseAndLogError(result.Handler, logger)
	logger.Debug().Msg("Document generated.")

	logger.Debug().Msgf("Generation: %v", time.Since(start))
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
		logger.Err(err).Msg("Sending metadata failed.")
		return status.Error(codes.Internal, "Sending metadata failed.")
	}

	chunks := make([]byte, 0, chunkSize)
	for {
		toRead := chunks[:cap(chunks)]
		numberOfReadBytes, err := result.Reader.Read(toRead)
		if numberOfReadBytes > 0 {
			chunks = chunks[:numberOfReadBytes]
			if err := stream.Send(&documentServiceApi.GeneratePreviewResponse{
				File: &documentServiceApi.GeneratePreviewResponse_Chunk{
					Chunk: chunks,
				},
			}); err != nil {
				logger.Err(err).Msg("Sending chunk failed.")
				return status.Error(codes.Internal, "Sending chunk failed.")
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Err(err).Msg("Generation of document failed.")
			return status.Error(codes.Internal, "generation of document failed.")
		}
	}

	logger.Debug().Msgf("Sending: %v", time.Since(start))
	return nil
}

func CloseAndLogError(fileHandler FileHandler, logger zerolog.Logger) {
	if err := fileHandler.Close(); err != nil {
		logger.Err(err).Msg("Closing and cleap files fail.")
	}
}
