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

	"github.com/kinneko-de/protobuf-go/encoding/protolua"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DocumentGeneratorLuatex struct {
}

func (DocumentGeneratorLuatex) GenerateDocument(requestId uuid.UUID, documentType string, message proto.Message) (result GeneratedFile, err error) {
	appRootDirectory := app.Config.RootPath
	luatexTemplateDirectory := path.Join(appRootDirectory, "template")
	runDirectory := path.Join(appRootDirectory, "run")
	tmpDirectory := path.Join(runDirectory, requestId.String())
	outputDirectoryRelativeToTmpDirectory := "generated"
	outputDirectory := path.Join(tmpDirectory, outputDirectoryRelativeToTmpDirectory)

	CreateDirectoryForRun(outputDirectory)

	documentInputData, err := convertToLuaTable(message)
	if err != nil {
		return result, err
	}

	templateFile, err := copyLuatexTemplate(luatexTemplateDirectory, documentType, tmpDirectory)
	if err != nil {
		return result, err
	}

	if err := createDocumentInputData(documentType, tmpDirectory, documentInputData); err != nil {
		return result, err
	}

	if err := executeLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory); err != nil {
		return result, err
	}
	if err := executeLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory); err != nil {
		return result, err
	}

	generatedDocumentFile, reader, err := createAccessToOutputfile(outputDirectory, documentType)
	if err != nil {
		return result, err
	}

	fileInfo, err := generatedDocumentFile.Stat()
	if err != nil {
		return result, err
	}

	return GeneratedFile{
		Reader: reader,
		Size:   fileInfo.Size(),
		Handler: GeneratedFileHandler{
			file:         generatedDocumentFile,
			tmpDirectory: tmpDirectory,
		},
		DocumentType: documentType,
	}, nil
}

func createAccessToOutputfile(outputDirectory string, documentType string) (*os.File, *bufio.Reader, error) {
	generatedDocument := path.Join(outputDirectory, documentType+".pdf")
	generatedDocumentFile, err := os.Open(generatedDocument)
	if err != nil {
		return nil, nil, fmt.Errorf("error open generated document %v: %v", generatedDocument, err)
	}
	reader := bufio.NewReader(generatedDocumentFile)
	return generatedDocumentFile, reader, nil
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
