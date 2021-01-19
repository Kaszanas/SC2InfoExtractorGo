package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"path/filepath"
	"time"
)

func initBufferWriter() (*bytes.Buffer, *zip.Writer) {

	// Create a buffer to write our archive to:
	buf := new(bytes.Buffer)
	// Create a new zip archive:
	w := zip.NewWriter(buf)

	return buf, w
}

func saveFileToArchive(replayString string, replayFile string, compressionMethod uint16, writer *zip.Writer) {
	jsonBytes := []byte(replayString)
	_, fileHeaderFilename := filepath.Split(replayFile)

	fh := &zip.FileHeader{
		Name:               filepath.Base(fileHeaderFilename) + ".json",
		UncompressedSize64: uint64(len(jsonBytes)),
		Method:             compressionMethod,
		Modified:           time.Now(),
	}
	fh.SetMode(0777)
	fw, err := writer.CreateHeader(fh)

	if err != nil {
		fmt.Printf("Error: %s", err)
		panic("Error")
	}

	fw.Write(jsonBytes)
}
