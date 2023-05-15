package operation

import (
	"log"
	"os"
)

func UseLogFileInGenerated() *os.File {
	createDirectory("log")
	f, err := os.OpenFile("log/log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	log.SetOutput(f)
	return f
}

func CloseLogFile(logfile *os.File) {
	err := logfile.Close()
	if err != nil {
		log.Fatalf("error closing log : %v", err)
	}
}

func createDirectory(directory string) {
	mkDirError := os.MkdirAll(directory, 0755)
	if mkDirError != nil {
		log.Fatalf("Can not create directory '%v': %v", directory, mkDirError)
	}
}
