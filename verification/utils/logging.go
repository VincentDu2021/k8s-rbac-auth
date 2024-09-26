package utils

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	FatalLogger *log.Logger
)

func Set_logging(path string) error {
	file, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime)
	FatalLogger = log.New(file, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)

	return err
}
