package main

import (
	"fmt"
	// "github.com/icza/mpq"
	"encoding/json"
	// "github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"strconv"
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

	PIDPlayerDescMap := replayFile.TrackerEvts.PIDPlayerDescMap
	// ToonPlayerDescMap := replayFile.TrackerEvts.ToonPlayerDescMap

	// Creating lists of strings for later use in generating JSON out of the replay data:
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

	// This structure is handled differently as it is a Map without .String() method:
	var PIDPlayerDescMapStrings []string
	for PIDPlayerDescKey, PIDPlayerDescValue := range PIDPlayerDescMap {

		// Converting ID to string:
		playerNumber := strconv.FormatInt(PIDPlayerDescKey, 10)

		// Converting struct to JSON:
		playerDescInformation, err := json.Marshal(PIDPlayerDescValue)

		if err != nil {
			panic(err)
		}

		// Putting everything together:
		PIDPlayerDescMapStrings = append(PIDPlayerDescMapStrings, "\""+playerNumber+"\":"+string(playerDescInformation))
	}

	// var ToonPlayerDescMapStrings []string
	// for ToonPlayerDescKey, TonPlayerDescValue := range PIDPlayerDescMap {

	// TODO: Check what are those events and how to get them to JSON format:
	gameEvtsErr := replayFile.GameEvtsErr
	messageEvtsErr := replayFile.MessageEvtsErr
	trackerEvtsErr := replayFile.TrackerEvtsErr

	// Crezting JSON structure by hand:
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
	fmt.Fprintf(&strBuilder, "  \"PIDPlayerDescMap\" : {%s}\n", strings.Join(PIDPlayerDescMapStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"")
	fmt.Fprintf(&strBuilder, "}")

	// Writing JSON file:
	_ = ioutil.WriteFile("./DEMOS/Output/11506446_1566325366_8429955.json", []byte(strBuilder.String()), 0644)

}
