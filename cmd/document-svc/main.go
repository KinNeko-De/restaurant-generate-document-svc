package main

import (
	"log"
	"net/http"

	"github.com/KinNeko-De/restaurant-document-svc/internal/app/operation"
)

func main() {
	logfile := operation.UseLogFileInGenerated()
	defer operation.CloseLogFile(logfile)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
