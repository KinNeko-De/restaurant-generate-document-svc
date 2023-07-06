package testing

import (
	"log"
	"time"

	"github.com/KinNeko-De/restaurant-document-svc/internal/app/document"
	"github.com/google/uuid"
	"github.com/kinneko-de/api-contract/golang/kinnekode/protobuf"
	restaurantApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant"
	restaurantdocumentApi "github.com/kinneko-de/api-contract/golang/kinnekode/restaurant/document"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GenerateTestInvoice(appRootDirectory string) {
	document.DocumentGenerator{}.GenerateDocument(createTestCommand(), appRootDirectory)
}

func createTestCommand() *restaurantdocumentApi.GenerateDocumentV1 {
	randomRequestId := createRandomUuid()

	request := restaurantdocumentApi.GenerateDocumentV1{
		Request: &restaurantApi.RequestV1{
			RequestId: randomRequestId,
		},
		RequestedDocuments: []*restaurantdocumentApi.GenerateDocumentV1_Document{
			{
				Type: &restaurantdocumentApi.GenerateDocumentV1_Document_Invoice{
					Invoice: &restaurantdocumentApi.InvoiceV1{
						DeliveredOn:  timestamppb.New(time.Date(2020, time.April, 13, 0, 0, 0, 0, time.UTC)),
						CurrencyCode: "EUR",
						Recipient: &restaurantdocumentApi.InvoiceV1_Recipient{
							Name:     "Max Mustermann",
							Street:   "Musterstraße 17",
							City:     "Musterstadt",
							PostCode: "12345",
							Country:  "DE",
						},
						Items: []*restaurantdocumentApi.InvoiceV1_Item{
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
				OutputFormats: []restaurantdocumentApi.GenerateDocumentV1_Document_OutputFormat{
					restaurantdocumentApi.GenerateDocumentV1_Document_OUTPUT_FORMAT_PDF,
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
