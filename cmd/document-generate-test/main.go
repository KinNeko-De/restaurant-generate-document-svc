package main

import (
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation"
	"github.com/kinneko-de/restaurant-document-generate-svc/internal/testing"
	"github.com/rs/zerolog"
)

func main() {
	operation.SetLoggingLevel(zerolog.DebugLevel)
	testing.GenerateTestInvoice()
}
