package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

// thisFileSuffix is this file's own path relative to the project root. Every
// source file compiled into the binary has the same absolute prefix baked in
// by the compiler, so stripping this known suffix from this file's own
// reported path yields that shared prefix.
const thisFileSuffix = "utils/logging_utils.go"

// projectRoot is the absolute path prefix embedded in the compiled binary
// that precedes every one of this project's source files. It is derived from
// this file's own reported path, so it works regardless of the machine or
// directory the binary was built in and needs no special build flags such as
// -trimpath.
var projectRoot = func() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return strings.TrimSuffix(filepath.ToSlash(file), thisFileSuffix)
}()

// callerPrettyfier trims the compile-time-embedded absolute source path down
// to a path relative to the project root, so logs don't leak the absolute
// path of the machine the binary was built on.
func callerPrettyfier(f *runtime.Frame) (function string, file string) {
	filePath := filepath.ToSlash(f.File)
	if projectRoot != "" {
		filePath = strings.TrimPrefix(filePath, projectRoot)
	}
	return f.Function, fmt.Sprintf("%s:%d", filePath, f.Line)
}

// setLogging contains logic that is used to initialize
// logging to a specified file with a specified level.
func SetLogging(logPath string, logLevel int) (*os.File, bool) {

	logDirectoryString := logPath
	log.SetFormatter(&log.JSONFormatter{CallerPrettyfier: callerPrettyfier})
	log.SetReportCaller(true)

	// Check if the directory exists:
	if _, err := os.Stat(logDirectoryString); os.IsNotExist(err) {
		log.WithField("error", err).
			Warn("Log directory does not exist. Creating it.")

		err := os.MkdirAll(logDirectoryString, 0755)
		if err != nil {
			log.WithField("error", err).Fatal("Cannot create log directory.")
			return &os.File{}, false
		}
	}

	// If the file doesn't exist, create it or append to the file
	logFileFilepath := logDirectoryString + "main_log.log"
	logFile, err := os.OpenFile(logFileFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
		return &os.File{}, false
	}

	log.SetOutput(logFile)
	log.Info("Set logging format, defined log file.")

	logLevelString := log.Level(logLevel).String()
	log.Info("Log level set to: " + logLevelString)

	log.SetLevel(log.Level(logLevel))
	log.Info("Set logging level.")

	return logFile, true
}
