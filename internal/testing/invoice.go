package testing

import (
	"bufio"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/google/uuid"
	apiProtobuf "github.com/kinneko-de/api-contract/golang/kinnekode/protobuf"
	apiRestaurantDocument "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document/v1"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/document"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GenerateTestInvoice() {
	appRootDirectory := app.Config.RootPath
	requestedDocument := createTestDocument()

	requestId := uuid.New()
	outputDirectory := path.Join(appRootDirectory, "output")
	document.CreateDirectoryForRun(outputDirectory)
	outputFile := path.Join(outputDirectory, requestId.String()+".pdf")
	f, err := os.Create(outputFile)
	if err != nil {
		operation.Logger.Fatal().Err(err).Msgf("Could not output file: %v", outputFile)
	}
	defer f.Close()
	testWriter := bufio.NewWriter(f)

	const chunkSize = 1000
	chunks := make([]byte, 0, chunkSize)
	result, err := document.DocumentGeneratorLuatex{}.GenerateDocument(requestId, requestedDocument)
	if err != nil {
		operation.Logger.Fatal().Err(err).Msg("Generation of document failed.")
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
			operation.Logger.Fatal().Err(err).Msg("Reading of document failed.")
		}
	}

	testWriter.Flush()
	operation.Logger.Info().Msgf("%v Bytes read", strconv.Itoa(totalReadBytes))

	if err := result.Handler.Close(); err != nil {
		operation.Logger.Fatal().Err(err).Msg("Closing of document failed.")
	}
}

func createTestDocument() *apiRestaurantDocument.RequestedDocument {
	request := &apiRestaurantDocument.RequestedDocument{
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
	}

	return request
}
