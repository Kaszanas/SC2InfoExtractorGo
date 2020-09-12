package main

import (
	"fmt"
	// "github.com/icza/mpq"
	// "github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	// "log"
)

func anonymize(listOfFiles []string) {

	// For every file in list of files open it and get the data inside:
	for _, file := range listOfFiles {
		replayFile, err := rep.NewFromFile(file)
		if err != nil {
			fmt.Printf("Failed to open file: %v\n", err)
			return
		}
		defer replayFile.Close()
	}
	// Receive replay object and anonymize the important data

	fmt.Println("Test")

}
