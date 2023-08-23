package document

import (
	"bufio"
	"io"
	"os"

	"github.com/google/uuid"
	restaurantDocumentApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
)

type DocumentGenerator interface {
	GenerateDocument(requestId uuid.UUID, command *restaurantDocumentApi.RequestedDocument) (result GeneratedFile, err error)
}

var (
	documentGenerator DocumentGenerator
)

func init() {
	documentGenerator = DocumentGeneratorLuatex{}
}

type GeneratedFile struct {
	Reader  *bufio.Reader
	Size    int64
	Handler Handler
}

type Handler interface {
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
	mkDirError := os.MkdirAll(outputDirectory, os.ModeExclusive)
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
