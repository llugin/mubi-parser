package debuglog

import (
	"log"
	"os"
)

const logFile = "/tmp/mubi.log"

var logger *log.Logger

func init() {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(f, "", log.Ldate|log.Lshortfile|log.Ltime)
}

// GetLogger returns reference to debug logger
func GetLogger() *log.Logger {
	return logger
}
