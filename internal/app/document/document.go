package document

import (
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

type DocumentGenerator struct {
}

func (documentGenerator DocumentGenerator) GenerateDocument(command *restaurantApi.GenerateDocument, appRootDirectory string) {
	luatexTemplateDirectory := path.Join(appRootDirectory, "template")

	runDirectory := path.Join(appRootDirectory, "run")

	tmpDirectory := path.Join(runDirectory, command.Request.RequestId.Value)
	outputDirectoryRelativeToTmpDirectory := "generated"
	outputDirectory := path.Join(tmpDirectory, outputDirectoryRelativeToTmpDirectory)
	documentGenerator.createDirectoryForRun(outputDirectory)

	template, message := documentGenerator.GetTemplateName(command)
	documentInputData := ToLuaTable(message)

	templateFile := documentGenerator.CopyLuatexTemplate(luatexTemplateDirectory, template, tmpDirectory)
	documentGenerator.CreateDocumentInputData(template, tmpDirectory, documentInputData)

	documentGenerator.ExecuteLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory)
	documentGenerator.ExecuteLuaLatex(outputDirectoryRelativeToTmpDirectory, templateFile, tmpDirectory)
	log.Println("Document generated.") // TODO make this debug
}

func (documentGenerator DocumentGenerator) ExecuteLuaLatex(outputDirectory string, templateFile string, tmpDirectory string) {
	cmd, commandError := documentGenerator.runCommand(outputDirectory, templateFile, tmpDirectory)

	if commandError != nil {
		log.Fatalf("error executing %v %v", cmd, commandError)
	}
}

func (documentGenerator DocumentGenerator) runCommand(outputDirectory string, templateFile string, tmpDirectory string) (*exec.Cmd, error) {
	outputParameter := "-output-directory=" + outputDirectory
	cmd := exec.Command("lualatex", outputParameter, templateFile)
	cmd.Dir = tmpDirectory
	commandError := cmd.Run()
	return cmd, commandError
}

func (documentGenerator DocumentGenerator) GetTemplateName(command *restaurantApi.GenerateDocument) (string, proto.Message) {
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

func (documentGenerator DocumentGenerator) CopyLuatexTemplate(documentDirectory string, template string, tmpDirectory string) string {
	templateFile := template + ".tex"
	_, texErr := copyFile(path.Join(documentDirectory, templateFile), path.Join(tmpDirectory, templateFile))
	if texErr != nil {
		log.Fatalf("Can not copy tex file: %v", texErr)
	}
	return templateFile
}

func (documentGenerator DocumentGenerator) CreateDocumentInputData(template string, tmpDirectory string, inputData []byte) {
	inputDataFile := "data.lua"
	file, err := os.Create(path.Join(tmpDirectory, inputDataFile))
	if err != nil {
		log.Fatalf("Error creating input data: %v", err)
	}
	file.WriteString("local ")
	file.Write(inputData)
	tableAssign := "return {" + strings.ToLower(template) + " = " + template + " }"
	file.WriteString(tableAssign)
	file.Close()
}

func (DocumentGenerator) createDirectoryForRun(outputDirectory string) {
	mkDirError := os.MkdirAll(outputDirectory, os.ModeExclusive)
	if mkDirError != nil {
		log.Fatalf("Can not create output directory: %v", mkDirError)
	}
}

func (DocumentGenerator) getCurrentDirectory() string {
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

func ToLuaTable(m proto.Message) []byte {
	opt := protolua.LuaMarshalOption{AdditionalMarshalers: []interface {
		Handle(fullName protoreflect.FullName) (protolua.MarshalFunc, error)
	}{protoluaextension.KinnekoDeProtobuf{}}}
	luaTable, err := opt.Marshal(m)
	if err != nil {
		log.Fatalf("Error converting protobuf message to luat table: %v", err)
	}
	return luaTable
}
