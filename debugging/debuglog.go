package debugging

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

const logFile = "mubi.log"

var (
	logger  *log.Logger
	created = false
)

// Log returns reference to debug logger
func Log() *log.Logger {
	return logger
}

// InitLogger initializes new logger
func InitLogger(logPath string, stderr bool) {
	var f io.Writer
	var err error
	if stderr {
		f = os.Stderr
	} else {
		f, err = os.OpenFile(filepath.Join(logPath, logFile),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	logger = log.New(f, "", log.Ldate|log.Lshortfile|log.Ltime)
}
