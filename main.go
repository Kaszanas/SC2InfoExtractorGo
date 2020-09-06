package main

import (
	"fmt"
	// "github.com/icza/mpq"
	// "github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"log"
	"os"
)

func listFiles() []os.FileInfo {

	files, err := ioutil.ReadDir("./Demos/Input")
	if err != nil {
		log.Fatal(err)
	}

	return files

}

func main() {

	testListFiles := listFiles()

	fmt.Println(testListFiles)

	for _, file := range testListFiles {
		fmt.Println(file.Name())
	}

	r, err := rep.NewFromFile("./DEMOS/Input/11506446_1566325366_8429955.SC2Replay")
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer r.Close()

	gameEventNames := map[string]bool{}

	for _, myString := range r.GameEvts {
		gameEventNames[myString.EvtType.Name] = true
		// fmt.Println(myString)
	}

	trackerEvents := r.TrackerEvts
	fmt.Printf("Tracker events: %d\n", len(trackerEvents.Evts))
	// trackerEventNames := map[string]bool{}
	// for _, myString := range r.TrackerEvts {
	// 	trackerEventNames[myString.EvtType.Name] = true
	// }

	fmt.Println(gameEventNames)

	// fmt.Printf("Version:        %v\n", r.Header.VersionString())
	// fmt.Printf("Loops:          %d\n", r.Header.Loops())
	// fmt.Printf("Length:         %v\n", r.Header.Duration())
	// fmt.Printf("Map:            %s\n", r.Details.Title())
	// fmt.Printf("Game events:    %d\n", len(r.GameEvts))
	// fmt.Printf("Message events: %d\n", len(r.MessageEvts))
	//

	// fmt.Println("Players:")
	// for _, p := range r.Details.Players() {
	// 	fmt.Printf("\tName: %-20s, Race: %c, Team: %d, Result: %v\n",
	// 		p.Name, p.Race().Letter, p.TeamID()+1, p.Result())
	// }
	// fmt.Printf("Full Header:\n%v\n", r.Header)

}
