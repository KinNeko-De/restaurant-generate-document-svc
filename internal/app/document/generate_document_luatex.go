package document

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app"
	protoluaextension "github.com/kinneko-de/restaurant-document-generate-svc/internal/app/encoding/protolua"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/logger"
	"github.com/rs/zerolog"

	"github.com/kinneko-de/protobuf-go/encoding/protolua"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DocumentGeneratorLuatex struct {
}

func (DocumentGeneratorLuatex) GenerateDocument(requestId uuid.UUID, documentType string, message proto.Message) (GeneratedFile, error) {
	appRootDirectory := app.Config.RootPath
	luatexTemplateDirectory := path.Join(appRootDirectory, "template")
	runDirectory := path.Join(appRootDirectory, "run")
	tmpDirectory := path.Join(runDirectory, requestId.String())
	outputDirectoryRelativeToTmpDirectory := "generated"
	outputDirectory := path.Join(tmpDirectory, outputDirectoryRelativeToTmpDirectory)
	logger := logger.Logger.With().
		Str("requestId", requestId.String()).
		Str("appRootDirectory", appRootDirectory).
		Str("luatexTemplateDirectory", luatexTemplateDirectory).
		Str("runDirectory", runDirectory).
		Str("tmpDirectory", tmpDirectory).
		Str("outputDirectoryRelativeToTmpDirectory", outputDirectoryRelativeToTmpDirectory).
		Str("outputDirectory", outputDirectory).
		Str("documentType", documentType).
		Logger()

	logger.Debug().Msg("Generating document...")
	generatedFile, err := generateDocument(logger, message, documentType, outputDirectory, luatexTemplateDirectory, tmpDirectory, outputDirectoryRelativeToTmpDirectory)
	if err != nil {
		logger.Error().Err(err).Msg("generating document failed")
		return GeneratedFile{}, err
	}

	logger.Debug().
		Str("fileSize", fmt.Sprintf("%v", generatedFile.Size)).
		Msgf("Document generated")
	return generatedFile, nil
}

func generateDocument(logger zerolog.Logger, message protoreflect.ProtoMessage, documentType string, outputDirectory string, luatexTemplateDirectory string, tmpDirectory string, outputDirectoryRelativeToTmpDirectory string) (GeneratedFile, error) {
	err := CreateDirectoryForRun(outputDirectory)
	if err != nil {
		return GeneratedFile{}, err
	}
	logger.Trace().Msg("Directory for run created")

	documentInputData, err := convertToLuaTable(message)
	if err != nil {
		return GeneratedFile{}, err
	}
	logger.Trace().Msg("Input data converted to lua table")

	templateFile, err := copyLuatexTemplate(luatexTemplateDirectory, documentType, tmpDirectory)
	if err != nil {
		return GeneratedFile{}, err
	}
	logger.Trace().Msg("Template copied")

	if err := createDocumentInputData(documentType, tmpDirectory, documentInputData); err != nil {
		return GeneratedFile{}, err
	}
	logger.Trace().Msg("Input data created")

	// Latex has to be executed twice because of the table of contents, page numbers, etc.
	// TODO: make this configurable over the template to save some time
	if err := executeLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory); err != nil {
		return GeneratedFile{}, err
	}
	if err := executeLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory); err != nil {
		return GeneratedFile{}, err
	}
	logger.Trace().Msg("LuaLatex executed")

	generatedDocumentFile, reader, err := readGeneratedDocument(outputDirectory, documentType)
	if err != nil {
		return GeneratedFile{}, err
	}
	logger.Trace().Msg("generated document read")

	fileInfo, err := generatedDocumentFile.Stat()
	if err != nil {
		return GeneratedFile{}, err
	}
	logger.Trace().Msg("generated document stat read")

	generatedFile := GeneratedFile{
		Reader: reader,
		Size:   fileInfo.Size(),
		Handler: GeneratedFileHandler{
			file:         generatedDocumentFile,
			tmpDirectory: tmpDirectory,
		},
		DocumentType: documentType,
	}
	return generatedFile, nil
}

func readGeneratedDocument(outputDirectory string, documentType string) (*os.File, *bufio.Reader, error) {
	generatedDocument := path.Join(outputDirectory, documentType+".pdf")
	generatedFile, err := os.Open(generatedDocument)
	if err != nil {
		return nil, nil, fmt.Errorf("error open generated document %v: %v", generatedDocument, err)
	}
	reader := bufio.NewReader(generatedFile)
	return generatedFile, reader, nil
}

func executeLuaLatex(outputDirectory string, templateFile string, tmpDirectory string) error {
	outputParameter := "-output-directory=" + outputDirectory
	cmd := exec.Command("lualatex", outputParameter, templateFile)
	cmd.Dir = tmpDirectory
	commandError := cmd.Run()
	if commandError != nil {
		return fmt.Errorf("error executing %v %v", cmd, commandError)
	}
	return nil
}

func copyLuatexTemplate(documentDirectory string, template string, tmpDirectory string) (string, error) {
	templateFile := template + ".tex"
	_, texErr := copyFile(path.Join(documentDirectory, templateFile), path.Join(tmpDirectory, templateFile))
	if texErr != nil {
		return "", fmt.Errorf("can not copy tex file: %v", texErr)
	}
	return templateFile, nil
}

func createDocumentInputData(documentType string, tmpDirectory string, inputData []byte) error {
	inputDataFile := "data.lua"
	file, err := os.Create(path.Join(tmpDirectory, inputDataFile))
	if err != nil {
		return fmt.Errorf("error creating input data to directory %v: %v", tmpDirectory, err)
	}
	file.WriteString("local ")
	file.Write(inputData)
	// TODO change protobuf-go to user lower names
	tableAssign := "return {" + strings.ToLower(documentType) + " = " + documentType + " }"
	file.WriteString(tableAssign)
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

func convertToLuaTable(m proto.Message) ([]byte, error) {
	opt := protolua.LuaMarshalOption{AdditionalMarshalers: []interface {
		Handle(fullName protoreflect.FullName) (protolua.MarshalFunc, error)
	}{protoluaextension.KinnekoDeProtobuf{}}}
	luaTable, err := opt.Marshal(m)
	if err != nil {
		err = fmt.Errorf("error converting protobuf message '%v' to luatable: %v", m, err)
	}
	return luaTable, err
}
