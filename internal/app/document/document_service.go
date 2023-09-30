package document

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	documentServiceApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"github.com/rs/zerolog"
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

	err := validateRequest(request)
	if err != nil {
		return err
	}

	requestId := uuid.New()
	logger := logger.Logger.With().Str("requestId", requestId.String()).Logger()

	logger.Debug().Msgf("Preprocessing finished: %v", time.Since(start))
	start = time.Now()

	document, err := GenerateDocument(requestId, request.RequestedDocument)
	if err != nil {
		return status.Error(codes.Internal, "Generation of document failed.")
	}
	defer CloseAndLogError(document.Handler, logger)

	logger.Debug().Msgf("Document generation finished: %v", time.Since(start))
	start = time.Now()

	err = sendMetadata(document, stream)
	if err != nil {
		return sendError(stream, err, logger, "Sending metadata failed.")
	}

	err = sendChuncks(document, stream)
	if err != nil {
		return sendError(stream, err, logger, "Sending chunk failed.")
	}

	logger.Debug().Msgf("Sending finished: %v", time.Since(start))
	metric.DocumentDelivered(document.DocumentType)

	return nil
}

func sendError(stream documentServiceApi.DocumentService_GeneratePreviewServer, err error, logger zerolog.Logger, messageTemplate string) error {
	if stream.Context().Err() == context.Canceled {
		logger.Debug().Msg("Stream canceled.")
		return nil
	}

	logger.Err(err).Msg(messageTemplate)
	return status.Error(codes.Internal, messageTemplate)
}

func validateRequest(request *documentServiceApi.GeneratePreviewRequest) error {
	if request.RequestedDocument == nil {
		return status.Error(codes.InvalidArgument, "requested document is mandatory to generate a document.")
	}
	return nil
}

func sendMetadata(document GeneratedFile, stream documentServiceApi.DocumentService_GeneratePreviewServer) error {
	return stream.Send(&documentServiceApi.GeneratePreviewResponse{
		File: &documentServiceApi.GeneratePreviewResponse_Metadata{
			Metadata: &documentServiceApi.GeneratedFileMetadata{
				CreatedAt: timestamppb.Now(),
				Size:      uint64(document.Size),
				MediaType: "application/pdf",
				Extension: ".pdf",
			},
		},
	})
}

func sendChuncks(document GeneratedFile, stream documentServiceApi.DocumentService_GeneratePreviewServer) error {
	chunks := make([]byte, 0, chunkSize)
	for {
		toRead := chunks[:cap(chunks)]
		numberOfReadBytes, err := document.Reader.Read(toRead)
		// If the error is EOF, it means that there is no more data to read from the file. We can return nil to indicate that there is no error.
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}
		if numberOfReadBytes == 0 {
			return status.Error(codes.Internal, "Reading file failed. Number of read bytes is 0 but no EndOfFile error returned.")
		}

		chunks = chunks[:numberOfReadBytes]
		err = sendReadBytes(chunks, stream)
		if err != nil {
			return err
		}
	}
}

func sendReadBytes(chunks []byte, stream documentServiceApi.DocumentService_GeneratePreviewServer) error {
	return stream.Send(&documentServiceApi.GeneratePreviewResponse{
		File: &documentServiceApi.GeneratePreviewResponse_Chunk{
			Chunk: chunks,
		},
	})
}

func CloseAndLogError(fileHandler FileHandler, logger zerolog.Logger) {
	if err := fileHandler.Close(); err != nil {
		logger.Err(err).Msg("Closing and cleanup generating files fail.")
	}
}
