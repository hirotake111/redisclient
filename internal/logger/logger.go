package logger

import (
	"log"
	"os"
)

const logFileName = "red.log"

// LogFilePath returns the full path to the log file in the system temp directory.
func LogFilePath() string {
	return "/tmp/" + logFileName
}

// Initialize sets the standard logger output to the log file in the temp directory.
func Initialize() error {
	logFile, err := os.OpenFile(LogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	log.SetOutput(logFile)
	return nil
}
