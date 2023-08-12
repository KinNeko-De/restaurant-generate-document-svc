package document

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app"
	protoluaextension "github.com/kinneko-de/restaurant-document-generate-svc/internal/app/encoding/protolua"

	restaurantDocumentApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/protobuf-go/encoding/protolua"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func GenerateDocument(requestId uuid.UUID, command *restaurantDocumentApi.RequestedDocument) (result GenerationResult, err error) {
	appRootDirectory := app.Config.RootPath
	luatexTemplateDirectory := path.Join(appRootDirectory, "template")
	runDirectory := path.Join(appRootDirectory, "run")
	tmpDirectory := path.Join(runDirectory, requestId.String())
	outputDirectoryRelativeToTmpDirectory := "generated"
	outputDirectory := path.Join(tmpDirectory, outputDirectoryRelativeToTmpDirectory)

	CreateDirectoryForRun(outputDirectory)

	rootObject, message := getTemplateName(command)
	documentInputData, err := convertToLuaTable(message)
	if err != nil {
		return result, err
	}

	templateFile, err := copyLuatexTemplate(luatexTemplateDirectory, rootObject, tmpDirectory)
	if err != nil {
		return result, err
	}

	if err := createDocumentInputData(rootObject, tmpDirectory, documentInputData); err != nil {
		return result, err
	}

	if err := executeLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory); err != nil {
		return result, err
	}
	if err := executeLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory); err != nil {
		return result, err
	}

	log.Println("Document generated.") // TODO make this debug

	generatedDocumentFile, reader, err := createAccessToOutputfile(outputDirectory, rootObject)
	if err != nil {
		return result, err
	}

	fileInfo, err := generatedDocumentFile.Stat()
	if err != nil {
		return result, err
	}

	return GenerationResult{
		generatedFile: generatedDocumentFile,
		tmpDirectory:  tmpDirectory,
		Reader:        reader,
		Size:          fileInfo.Size(),
	}, nil
}

func createAccessToOutputfile(outputDirectory string, rootObject string) (*os.File, *bufio.Reader, error) {
	generatedDocument := path.Join(outputDirectory, rootObject+".pdf")
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

func getTemplateName(command *restaurantDocumentApi.RequestedDocument) (string, proto.Message) {
	var rootObject string
	var message proto.Message
	switch command.Type.(type) {
	case *restaurantDocumentApi.RequestedDocument_Invoice:
		rootObject = "Invoice"
		message = command.GetInvoice()
	default:
		log.Fatalf("Document %v not supported yet", command.Type)
	}
	return rootObject, message
}

func copyLuatexTemplate(documentDirectory string, template string, tmpDirectory string) (string, error) {
	templateFile := template + ".tex"
	_, texErr := copyFile(path.Join(documentDirectory, templateFile), path.Join(tmpDirectory, templateFile))
	if texErr != nil {
		return "", fmt.Errorf("can not copy tex file: %v", texErr)
	}
	return templateFile, nil
}

func createDocumentInputData(rootObject string, tmpDirectory string, inputData []byte) error {
	inputDataFile := "data.lua"
	file, err := os.Create(path.Join(tmpDirectory, inputDataFile))
	if err != nil {
		return fmt.Errorf("error creating input data to directory %v: %v", tmpDirectory, err)
	}
	file.WriteString("local ")
	file.Write(inputData)
	// TODO change protobuf-go to user lower names
	tableAssign := "return {" + strings.ToLower(rootObject) + " = " + rootObject + " }"
	file.WriteString(tableAssign)
	if err := file.Close(); err != nil {
		return err
	}
	return nil
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

type GenerationResult struct {
	generatedFile *os.File
	tmpDirectory  string
	Reader        *bufio.Reader
	Size          int64
}

func (generationResult GenerationResult) Close() error {
	generationResult.generatedFile.Close()
	err := os.RemoveAll(generationResult.tmpDirectory)
	return err
}
