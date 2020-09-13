package main

import (
	"fmt"
	// "github.com/icza/mpq"
	// "encoding/json"
	// "github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"strings"
	// "log"
)

func main() {

	// Function defined in path_utils
	// Getting list of files within a directory:
	testListFiles := listFiles("./DEMOS/Input")

	for _, file := range testListFiles {
		fmt.Println(file.Name())
	}

	replayFilepath := "./DEMOS/Input/11506446_1566325366_8429955.SC2Replay"

	replayFile, err := rep.NewFromFile(replayFilepath)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer replayFile.Close()

	header := replayFile.Header.String()
	details := replayFile.Details.String()
	initData := replayFile.InitData.String()
	attrEvts := replayFile.AttrEvts.String()
	metadata := replayFile.Metadata.String()

	var gameEventStrings []string
	for _, gameEvent := range replayFile.GameEvts {
		gameEventStrings = append(gameEventStrings, gameEvent.String())
	}

	var messageEventStrings []string
	for _, messageEvent := range replayFile.MessageEvts {
		messageEventStrings = append(messageEventStrings, messageEvent.String())
	}

	var trackerEventStrings []string
	for _, trackerEvent := range replayFile.TrackerEvts.Evts {
		trackerEventStrings = append(trackerEventStrings, trackerEvent.String())
	}

	// PIDDescMap := replayFile.TrackerEvts.PIDPlayerDescMap
	// ToonDescMap := replayFile.TrackerEvts.ToonPlayerDescMap

	gameEvtsErr := replayFile.GameEvtsErr
	messageEvtsErr := replayFile.MessageEvtsErr
	trackerEvtsErr := replayFile.TrackerEvtsErr

	var strBuilder strings.Builder
	fmt.Fprintf(&strBuilder, "{\n")
	fmt.Fprintf(&strBuilder, "  \"header\" : %s,\n", header)
	fmt.Fprintf(&strBuilder, "  \"initData\" : %s,\n", initData)
	fmt.Fprintf(&strBuilder, "  \"details\" : %s,\n", details)
	fmt.Fprintf(&strBuilder, "  \"attrEvts\" : %s,\n", attrEvts)
	fmt.Fprintf(&strBuilder, "  \"metadata\" : %s,\n", metadata)
	fmt.Fprintf(&strBuilder, "  \"gameEvtsErr\" : %s\n", gameEvtsErr)
	fmt.Fprintf(&strBuilder, "  \"messageEvtsErr\" : %s\n", messageEvtsErr)
	fmt.Fprintf(&strBuilder, "  \"trackerEvtsErr\" : %s\n", trackerEvtsErr)
	fmt.Fprintf(&strBuilder, "  \"messageEventsStrings\" : [%s]\n", strings.Join(messageEventStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"gameEventStrings\" : [%s]\n", strings.Join(gameEventStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"trackerEventStrings\" : [%s]\n", strings.Join(trackerEventStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"")
	fmt.Fprintf(&strBuilder, "}")

	// stringifiedReplayData2 := "{\n" + "\"header\": " + header + "}"

	// jsonFile, _ := json.Marshal(stringifiedReplayData2)

	_ = ioutil.WriteFile("./DEMOS/Output/11506446_1566325366_8429955.json", []byte(strBuilder.String()), 0644)

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
