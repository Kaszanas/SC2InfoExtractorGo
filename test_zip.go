package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/larzconwell/bzip2"
	"io"
	"io/ioutil"
	"time"
)

func main() {

	// Register a custom compressor:
	zip.RegisterCompressor(12, func(out io.Writer) (io.WriteCloser, error) {
		return bzip2.NewWriterLevel(out, 9)
	})

	// Create a buffer to write our archive to:
	buf := new(bytes.Buffer)

	// Create a new zip archive:
	w := zip.NewWriter(buf)

	// Creating a sample string:
	someString := "Testing bzip2 compression algorithm"

	// Converting from string to bytes:
	someStringBytes := []byte(someString)

	fh := &zip.FileHeader{
		Name:               "test.txt",
		UncompressedSize64: uint64(len(someStringBytes)),
		Method:             12,
		Modified:           time.Now(),
	}
	fh.SetMode(0777)
	fw, err := w.CreateHeader(fh)

	if err != nil {
		fmt.Printf("Error: %s", err)
		panic("Error")
	}

	fw.Write(someStringBytes)
	w.Close()

	_ = ioutil.WriteFile("test_zip.zip", buf.Bytes(), 0777)
}
