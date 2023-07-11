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

	protoluaextension "github.com/KinNeko-De/restaurant-document-svc/internal/app/encoding/protolua"

	restaurantApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/protobuf-go/encoding/protolua"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func GenerateDocument(command *restaurantApi.GenerateDocument, appRootDirectory string) (result GenerationResult, err error) {
	luatexTemplateDirectory := path.Join(appRootDirectory, "template")
	runDirectory := path.Join(appRootDirectory, "run")
	tmpDirectory := path.Join(runDirectory, command.Request.RequestId.Value)
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

	return GenerationResult{
		generatedFile: generatedDocumentFile,
		tmpDirectory:  tmpDirectory,
		Reader:        reader,
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
	cmd, commandError := runCommand(outputDirectory, templateFile, tmpDirectory)
	if commandError != nil {
		return fmt.Errorf("error executing %v %v", cmd, commandError)
	}

	return nil
}

func runCommand(outputDirectory string, templateFile string, tmpDirectory string) (*exec.Cmd, error) {
	outputParameter := "-output-directory=" + outputDirectory
	cmd := exec.Command("lualatex", outputParameter, templateFile)
	cmd.Dir = tmpDirectory
	commandError := cmd.Run()
	return cmd, commandError
}

func getTemplateName(command *restaurantApi.GenerateDocument) (string, proto.Message) {
	var rootObject string
	var message proto.Message
	switch command.RequestedDocuments[0].Type.(type) {
	case *restaurantApi.RequestedDocument_Invoice:
		rootObject = "Invoice"
		message = command.RequestedDocuments[0].GetInvoice()
	default:
		log.Fatalf("Document %v not supported yet", command.RequestedDocuments[0].Type)
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
		err = fmt.Errorf("Error converting protobuf message '%v' to luatable: %v", m, err)
	}
	return luaTable, err
}

type GenerationResult struct {
	generatedFile *os.File
	tmpDirectory  string
	Reader        *bufio.Reader
}

func (generationResult GenerationResult) Close() error {
	generationResult.generatedFile.Close()
	err := os.RemoveAll(generationResult.tmpDirectory)
	return err
}
