package document

import (
	"bufio"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	restaurantDocumentApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"google.golang.org/protobuf/proto"
)

type DocumentGenerator interface {
	GenerateDocument(requestId uuid.UUID, documentType string, message proto.Message) (GeneratedFile, error)
}

var (
	documentGenerator DocumentGenerator = DocumentGeneratorLuatex{}
)

func GenerateDocument(requestId uuid.UUID, requestedDocument *restaurantDocumentApi.RequestedDocument) (GeneratedFile, error) {
	documentType, message := parseRequest(requestedDocument)
	logger := logger.Logger.With().
		Str("requestId", requestId.String()).
		Str("documentType", documentType).
		Logger()

	start := time.Now()
	generatedFile, err := documentGenerator.GenerateDocument(requestId, documentType, message)
	metric.DocumentGenerated(documentType, time.Since(start), err)
	if err != nil {
		logger.Error().Err(err).Msg("generating document failed")
	} else {
		logger.Info().Dur("duration", time.Since(start)).Msg("document generated successfully")
	}

	return generatedFile, err
}

func parseRequest(command *restaurantDocumentApi.RequestedDocument) (string, proto.Message) {
	ref := command.ProtoReflect()
	refDescriptor := ref.Descriptor()
	setValue := ref.WhichOneof(refDescriptor.Oneofs().ByName("type"))
	fieldName := setValue.Message().Name()
	message := command.ProtoReflect().Get(setValue).Message().Interface()
	documentType := string(fieldName)
	return documentType, message
}

type GeneratedFile struct {
	Reader       *bufio.Reader
	Size         int64
	Handler      FileHandler
	DocumentType string
}

type FileHandler interface {
	Close() error
}

type GeneratedFileHandler struct {
	file         *os.File
	tmpDirectory string
}

func (generatedFileHandler GeneratedFileHandler) Close() error {
	closeErr := generatedFileHandler.file.Close()
	if closeErr != nil {
		return closeErr
	}
	err := os.RemoveAll(generatedFileHandler.tmpDirectory)
	return err
}

func CreateDirectoryForRun(outputDirectory string) error {
	mkDirError := os.MkdirAll(outputDirectory, os.FileMode(0770))
	return mkDirError
}

func copyFile(src, dst string) (int64, error) {
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
