package utils

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// initBufferWriter initializes a bytes buffer that is used to hold all of the information that writer will write to the archive
func InitBufferWriter() (*bytes.Buffer, *zip.Writer) {

	log.Info("Entered initBufferWriter()")

	// Create a buffer to write our archive to:
	buf := new(bytes.Buffer)
	// Create a new zip archive:
	w := zip.NewWriter(buf)

	log.Info("Finished initBufferWriter()")

	return buf, w
}

// SaveFileToArchive creates a file header and saves replayString (JSON) bytes into the zip writer
func SaveFileToArchive(replayString string, replayFile string, compressionMethod uint16, writer *zip.Writer) bool {

	log.Info("Entered saveFileToArchive()")

	jsonBytes := []byte(replayString)
	_, fileHeaderFilename := filepath.Split(replayFile)

	fh := &zip.FileHeader{
		Name:               filepath.Base(fileHeaderFilename) + ".json",
		UncompressedSize64: uint64(len(jsonBytes)),
		Method:             compressionMethod,
		Modified:           time.Now(),
	}
	fh.SetMode(0777)
	log.WithFields(log.Fields{
		"name":              fh.Name,
		"uncompressedSize":  fh.UncompressedSize64,
		"compressionMethod": fh.Method,
		"modified":          fh.Modified}).Debug("Created file header.")

	fw, err := writer.CreateHeader(fh)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  replayFile,
			"error": err}).Error("Got error when adding a file header to the archive.")
		return false
	}

	_, err = fw.Write(jsonBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"file":             replayFile,
			"error":            err,
			"compressionError": true}).Error("Got error when adding a file header to the archive.")
		return false
	}

	log.Info("Finished SaveFileToArchive()")

	return true
}

// SaveFileToDrive is a helper function that takes the json string of a StarCraft II replay and writes it to drive.
func SaveFileToDrive(replayString string, replayFile string, absolutePathOutputDirectory string) bool {

	_, replayFileNameWithExt := filepath.Split(replayFile)

	replayFileName := strings.TrimSuffix(replayFileNameWithExt, filepath.Ext(replayFileNameWithExt))

	jsonAbsPath := filepath.Join(absolutePathOutputDirectory, replayFileName+".json")
	jsonBytes := []byte(replayString)

	err := ioutil.WriteFile(jsonAbsPath, jsonBytes, 0777)
	if err != nil {
		log.WithField("replayFile", replayFile).Error("Failed to write .json to drive!")
		return false
	}

	return true
}
