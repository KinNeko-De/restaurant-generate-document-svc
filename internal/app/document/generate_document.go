package document

import (
	"bufio"
	"io"
	"os"

	"github.com/google/uuid"
	restaurantDocumentApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

type DocumentGenerator interface {
	GenerateDocument(requestId uuid.UUID, documentType string, message proto.Message) (result GeneratedFile, err error)
}

var (
	documentGenerator DocumentGenerator = DocumentGeneratorLuatex{}
)

func GenerateDocument(requestId uuid.UUID, requestedDocument *restaurantDocumentApi.RequestedDocument, logger zerolog.Logger) (result GeneratedFile, err error) {
	documentType, message := getTemplateName(requestedDocument)
	metric.DocumentRequested(documentType)

	generatedFile, err := documentGenerator.GenerateDocument(requestId, documentType, message)
	if err == nil {
		metric.DocumentGenerated(documentType)
	} else {
		logger.Err(err).Msg("Generation of document failed.")
		metric.DocumentFailed(documentType)
	}

	return generatedFile, err
}

func getTemplateName(command *restaurantDocumentApi.RequestedDocument) (string, proto.Message) {
	ref := command.ProtoReflect()
	refDescriptor := ref.Descriptor()
	setValue := ref.WhichOneof(refDescriptor.Oneofs().ByName("type"))
	fieldName := setValue.Message().Name()
	message := command.ProtoReflect().Get(setValue).Message().Interface()
	rootObject := string(fieldName)
	return rootObject, message
}

type GeneratedFile struct {
	Reader  *bufio.Reader
	Size    int64
	Handler FileHandler
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
