package main

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

func initBufferWriter() (*bytes.Buffer, *zip.Writer) {

	log.Info("Entered initBufferWriter()")
	// Create a buffer to write our archive to:
	buf := new(bytes.Buffer)
	// Create a new zip archive:
	w := zip.NewWriter(buf)

	return buf, w
}

func saveFileToArchive(replayString string, replayFile string, compressionMethod uint16, writer *zip.Writer) bool {

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
			"error": err}).Warn("Got error when adding a file header to the archive.")
		return false
	}

	fw.Write(jsonBytes)

	return true
}
