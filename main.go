package main

import (
	"fmt"
	// "github.com/icza/mpq"
	// "github.com/icza/s2prot"
	"encoding/json"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	// "log"
)

func main() {

	// Function defined in path_utils
	// Getting list of files within a directory:

	// testListFiles := listFiles("./DEMOS/Input")

	// for _, file := range testListFiles {
	// 	// fmt.Println(file.Name())
	// }

	replayFilepath := "./DEMOS/Input/11506446_1566325366_8429955.SC2Replay"

	replayFile, err := rep.NewFromFile(replayFilepath)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer replayFile.Close()

	gameEventNames := map[string]bool{}
	gameEvents := replayFile.GameEvts
	for _, gameEventObject := range gameEvents {
		gameEventNames[gameEventObject.EvtType.Name] = true
		// fmt.Println(myString)
	}

	trackerEvents := replayFile.TrackerEvts
	fmt.Printf("Tracker events: %d\n", len(trackerEvents.Evts))

	trackerEventNames := map[string]bool{}
	for _, event := range trackerEvents.Evts {

		if val, ok := trackerEventNames[event.EvtType.Name]; ok {
		} else {
			trackerEventNames[event.EvtType.Name] = val
			fmt.Println(event)
		}

	}

	fmt.Println(gameEventNames)
	fmt.Println(trackerEventNames)

	players := replayFile.Details.Players()

	for _, player := range players {
		fmt.Println(player)
	}

	jsonFile, _ := json.MarshalIndent(replayFile, "", " ")

	_ = ioutil.WriteFile("./DEMOS/Output/11506446_1566325366_8429955.json", jsonFile, 0644)

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
