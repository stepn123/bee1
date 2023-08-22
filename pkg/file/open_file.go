package file

import (
	"os"
	"path/filepath"
)

func OpenLogFile() (*os.File, error){
	logPath := filepath.Join("logs", "server.log")

	return os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
}