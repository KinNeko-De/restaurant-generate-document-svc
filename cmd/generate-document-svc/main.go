package main

import (
	"log"
	"net/http"

	"github.com/kinneko-de/restaurant-generate-document-svc/build"

	"github.com/kinneko-de/restaurant-generate-document-svc/internal/app/operation"
)

func main() {
	logfile := operation.UseLogFileInGenerated()
	defer operation.CloseLogFile(logfile)

	log.Println("Version " + build.Version)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
