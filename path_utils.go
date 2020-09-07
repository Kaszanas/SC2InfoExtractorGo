package sc2processing

import (
	"io/ioutil"
	"log"
	"os"
)

func listFiles(path string) []os.FileInfo {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	return files

}
