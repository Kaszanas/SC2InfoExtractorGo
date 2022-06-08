package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// setLogging contains logic that is used to initialize logging to a specified file with a specified level.
func SetLogging(logPath string, logLevel int) (*os.File, bool) {

	logDirectoryString := logPath
	log.SetFormatter(&log.JSONFormatter{})

	// If the file doesn't exist, create it or append to the file
	logFileFilepath := logDirectoryString + "main_log.log"
	logFile, err := os.OpenFile(logFileFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
		return &os.File{}, false
	}

	log.SetOutput(logFile)
	log.Info("Set logging format, defined log file.")

	log.SetLevel(log.Level(logLevel))
	log.Info("Set logging level.")

	return logFile, true

}
