package document

import (
	"bufio"
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

func GenerateDocument(command *restaurantApi.GenerateDocument, appRootDirectory string) (result GenerationResult) {
	luatexTemplateDirectory := path.Join(appRootDirectory, "template")
	runDirectory := path.Join(appRootDirectory, "run")
	tmpDirectory := path.Join(runDirectory, command.Request.RequestId.Value)
	outputDirectoryRelativeToTmpDirectory := "generated"
	outputDirectory := path.Join(tmpDirectory, outputDirectoryRelativeToTmpDirectory)

	CreateDirectoryForRun(outputDirectory)

	rootObject, message := getTemplateName(command)
	documentInputData := convertToLuaTable(message)

	templateFile := copyLuatexTemplate(luatexTemplateDirectory, rootObject, tmpDirectory)
	createDocumentInputData(rootObject, tmpDirectory, documentInputData)

	executeLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory)
	executeLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory)
	log.Println("Document generated.") // TODO make this debug

	generatedDocument := path.Join(outputDirectory, rootObject+".pdf")
	generatedDocumentFile, err := os.Open(generatedDocument)
	if err != nil {
		log.Fatalf("error open generated document %v", generatedDocument)
	}

	reader := bufio.NewReader(generatedDocumentFile)
	return GenerationResult{
		generatedFile: generatedDocumentFile,
		tmpDirectory:  tmpDirectory,
		Reader:        reader,
	}
}

func executeLuaLatex(outputDirectory string, templateFile string, tmpDirectory string) {
	cmd, commandError := runCommand(outputDirectory, templateFile, tmpDirectory)

	if commandError != nil {
		log.Fatalf("error executing %v %v", cmd, commandError)
	}
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
	case *restaurantApi.GenerateDocument_Document_Invoice:
		rootObject = "Invoice"
		message = command.RequestedDocuments[0].GetInvoice()
	default:
		log.Fatalf("Document %v not supported yet", command.RequestedDocuments[0].Type)
	}
	return rootObject, message
}

func copyLuatexTemplate(documentDirectory string, template string, tmpDirectory string) string {
	templateFile := template + ".tex"
	_, texErr := copyFile(path.Join(documentDirectory, templateFile), path.Join(tmpDirectory, templateFile))
	if texErr != nil {
		log.Fatalf("Can not copy tex file: %v", texErr)
	}
	return templateFile
}

func createDocumentInputData(rootObject string, tmpDirectory string, inputData []byte) {
	inputDataFile := "data.lua"
	file, err := os.Create(path.Join(tmpDirectory, inputDataFile))
	if err != nil {
		log.Fatalf("Error creating input data: %v", err)
	}
	file.WriteString("local ")
	file.Write(inputData)
	// TODO change protobuf-go to user lower names
	tableAssign := "return {" + strings.ToLower(rootObject) + " = " + rootObject + " }"
	file.WriteString(tableAssign)
	file.Close()
}

func CreateDirectoryForRun(outputDirectory string) {
	mkDirError := os.MkdirAll(outputDirectory, os.ModeExclusive)
	if mkDirError != nil {
		log.Fatalf("Can not create output directory: %v", mkDirError)
	}
}

func getCurrentDirectory() string {
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalf("error get current directory: %v", err)
	}
	return currentDirectory
}

func copyFile(src, dst string) (int64, error) {
	source, openError := os.Open(src)
	if openError != nil {
		return 0, openError
	}
	defer source.Close()

	destination, createError := os.Create(dst)
	if createError != nil {
		return 0, createError
	}
	defer destination.Close()
	nBytes, copyError := io.Copy(destination, source)
	return nBytes, copyError
}

func convertToLuaTable(m proto.Message) []byte {
	opt := protolua.LuaMarshalOption{AdditionalMarshalers: []interface {
		Handle(fullName protoreflect.FullName) (protolua.MarshalFunc, error)
	}{protoluaextension.KinnekoDeProtobuf{}}}
	luaTable, err := opt.Marshal(m)
	if err != nil {
		log.Fatalf("Error converting protobuf message to luat table: %v", err)
	}
	return luaTable
}

type GenerationResult struct {
	generatedFile *os.File
	tmpDirectory  string
	Reader        *bufio.Reader
}

func (generationResult GenerationResult) Close() {
	generationResult.generatedFile.Close()
	err := os.RemoveAll(generationResult.tmpDirectory)
	if err != nil {
		log.Fatalf("Deleting tmp directory failed: %v", err)
	}
}
