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
	apiProtobuf "github.com/kinneko-de/api-contract/golang/kinnekode/protobuf"
	apiRestaurantDocument "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	restaurantApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/document"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GenerateTestInvoice() {
	appRootDirectory := app.Config.RootPath
	testCommand := createTestCommand()

	requestId, err := apiProtobuf.ToUuid(testCommand.Request.GetRequestId())
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
	result, err := document.GenerateDocument(requestId, testCommand.RequestedDocuments[0])
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

func createTestCommand() *apiRestaurantDocument.GenerateDocument {
	randomRequestId := createRandomUuid()

	request := apiRestaurantDocument.GenerateDocument{
		Request: &restaurantApi.Request{
			RequestId: randomRequestId,
		},
		RequestedDocuments: []*apiRestaurantDocument.RequestedDocument{
			{
				Type: &apiRestaurantDocument.RequestedDocument_Invoice{
					Invoice: &apiRestaurantDocument.Invoice{
						DeliveredOn:  timestamppb.New(time.Date(2020, time.April, 13, 0, 0, 0, 0, time.UTC)),
						CurrencyCode: "EUR",
						Recipient: &apiRestaurantDocument.Invoice_Recipient{
							Name:     "Max Mustermann",
							Street:   "Musterstraße 17",
							City:     "Musterstadt",
							PostCode: "12345",
							Country:  "DE",
						},
						Items: []*apiRestaurantDocument.Invoice_Item{
							{
								Description: "Spitzenunterwäsche\r\nANS 23054303053",
								Quantity:    2,
								NetAmount:   &apiProtobuf.Decimal{Value: "3.35"},
								Taxation:    &apiProtobuf.Decimal{Value: "19"},
								TotalAmount: &apiProtobuf.Decimal{Value: "3.99"},
								Sum:         &apiProtobuf.Decimal{Value: "7.98"},
							},
							{
								Description: "Schlabberhose (10% reduziert)\r\nANS 606406540",
								Quantity:    1,
								NetAmount:   &apiProtobuf.Decimal{Value: "9.07"},
								Taxation:    &apiProtobuf.Decimal{Value: "19"},
								TotalAmount: &apiProtobuf.Decimal{Value: "10.79"},
								Sum:         &apiProtobuf.Decimal{Value: "10.79"},
							},
							{
								Description: "Versandkosten",
								Quantity:    1,
								NetAmount:   &apiProtobuf.Decimal{Value: "0.00"},
								Taxation:    &apiProtobuf.Decimal{Value: "0"},
								TotalAmount: &apiProtobuf.Decimal{Value: "0.00"},
								Sum:         &apiProtobuf.Decimal{Value: "0.00"},
							},
						},
					},
				},
				OutputFormats: []apiRestaurantDocument.RequestedDocument_OutputFormat{
					apiRestaurantDocument.RequestedDocument_OUTPUT_FORMAT_PDF,
				},
			},
		},
	}

	return &request
}

func createRandomUuid() *apiProtobuf.Uuid {
	id, uuidErr := uuid.NewUUID()
	if uuidErr != nil {
		log.Fatalf("error generating google uuid: %v", uuidErr)
	}
	randomRequestId, protobufErr := apiProtobuf.ToProtobuf(id)
	if protobufErr != nil {
		log.Fatalf("error generating protobuf uuid: %v", protobufErr)
	}
	return randomRequestId
}
