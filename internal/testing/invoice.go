package testing

import (
	"bufio"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kinneko-de/api-contract/golang/kinnekode/protobuf"
	restaurantDocumentApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	restaurantApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/v1"
	"github.com/kinneko-de/restaurant-generate-document-svc/internal/app/document"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GenerateTestInvoice(appRootDirectory string) {
	testCommand := createTestCommand()

	requestId, err := protobuf.ToUuid(testCommand.Request.GetRequestId())
	if err != nil {
		log.Fatal(err)
	}
	outputDirectory := path.Join(appRootDirectory, "output")
	document.CreateDirectoryForRun(outputDirectory)
	f, err := os.Create(path.Join(outputDirectory, requestId.String()+".pdf"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	testWriter := bufio.NewWriter(f)

	const chunkSize = 1000
	chunks := make([]byte, 0, chunkSize)
	result, err := document.GenerateDocument(requestId, testCommand.RequestedDocuments[0], appRootDirectory)
	if err != nil {
		log.Fatal(err)
	}
	totalReadBytes := 0
	for {
		numberOfReadBytes, err := result.Reader.Read(chunks[:cap(chunks)])
		if numberOfReadBytes > 0 {
			chunks = chunks[:numberOfReadBytes]
			totalReadBytes += len(chunks)
			testWriter.Write(chunks)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	testWriter.Flush()
	log.Println(strconv.Itoa(totalReadBytes) + "Bytes read")

	if err := result.Close(); err != nil {
		log.Fatal(err)
	}
}

func createTestCommand() *restaurantDocumentApi.GenerateDocument {
	randomRequestId := createRandomUuid()

	request := restaurantDocumentApi.GenerateDocument{
		Request: &restaurantApi.Request{
			RequestId: randomRequestId,
		},
		RequestedDocuments: []*restaurantDocumentApi.RequestedDocument{
			{
				Type: &restaurantDocumentApi.RequestedDocument_Invoice{
					Invoice: &restaurantDocumentApi.Invoice{
						DeliveredOn:  timestamppb.New(time.Date(2020, time.April, 13, 0, 0, 0, 0, time.UTC)),
						CurrencyCode: "EUR",
						Recipient: &restaurantDocumentApi.Invoice_Recipient{
							Name:     "Max Mustermann",
							Street:   "Musterstraße 17",
							City:     "Musterstadt",
							PostCode: "12345",
							Country:  "DE",
						},
						Items: []*restaurantDocumentApi.Invoice_Item{
							{
								Description: "Spitzenunterwäsche\r\nANS 23054303053",
								Quantity:    2,
								NetAmount:   &protobuf.Decimal{Value: "3.35"},
								Taxation:    &protobuf.Decimal{Value: "19"},
								TotalAmount: &protobuf.Decimal{Value: "3.99"},
								Sum:         &protobuf.Decimal{Value: "7.98"},
							},
							{
								Description: "Schlabberhose (10% reduziert)\r\nANS 606406540",
								Quantity:    1,
								NetAmount:   &protobuf.Decimal{Value: "9.07"},
								Taxation:    &protobuf.Decimal{Value: "19"},
								TotalAmount: &protobuf.Decimal{Value: "10.79"},
								Sum:         &protobuf.Decimal{Value: "10.79"},
							},
							{
								Description: "Versandkosten",
								Quantity:    1,
								NetAmount:   &protobuf.Decimal{Value: "0.00"},
								Taxation:    &protobuf.Decimal{Value: "0"},
								TotalAmount: &protobuf.Decimal{Value: "0.00"},
								Sum:         &protobuf.Decimal{Value: "0.00"},
							},
						},
					},
				},
				OutputFormats: []restaurantDocumentApi.RequestedDocument_OutputFormat{
					restaurantDocumentApi.RequestedDocument_OUTPUT_FORMAT_PDF,
				},
			},
			{},
		},
	}

	return &request
}

func createRandomUuid() *protobuf.Uuid {
	id, uuidErr := uuid.NewUUID()
	if uuidErr != nil {
		log.Fatalf("error generating google uuid: %v", uuidErr)
	}
	randomRequestId, protobufErr := protobuf.ToProtobuf(id)
	if protobufErr != nil {
		log.Fatalf("error generating protobuf uuid: %v", protobufErr)
	}
	return randomRequestId
}
