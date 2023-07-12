package main

import (
	"log"
	"net/http"

	root "github.com/kinneko-de/restaurant-generate-document-svc"

	"github.com/KinNeko-De/restaurant-generate-document-svc/internal/app/operation"
)

func main() {
	logfile := operation.UseLogFileInGenerated()
	defer operation.CloseLogFile(logfile)

	log.Println("Version " + root.Version)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
