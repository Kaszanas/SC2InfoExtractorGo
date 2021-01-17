package main

import (
	"encoding/json"
	"fmt"
	"github.com/icza/s2prot/rep"
	"strconv"
	"strings"
)

func stringifyReplay(replayFile string) (bool, string) {

	replayData, err := rep.NewFromFile(replayFile)

	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)

		return false, ""
	}
	defer replayData.Close()

	header := replayData.Header.String()
	details := replayData.Details.String()
	initData := replayData.InitData.String()
	attrEvts := replayData.AttrEvts.String()
	metadata := replayData.Metadata.String()

	PIDPlayerDescMap := replayData.TrackerEvts.PIDPlayerDescMap
	ToonPlayerDescMap := replayData.TrackerEvts.ToonPlayerDescMap

	// Creating lists of strings for later use in generating JSON out of the replay data:
	var gameEventStrings []string
	for _, gameEvent := range replayData.GameEvts {
		gameEventStrings = append(gameEventStrings, gameEvent.String())
	}

	var messageEventStrings []string
	for _, messageEvent := range replayData.MessageEvts {
		messageEventStrings = append(messageEventStrings, messageEvent.String())
	}

	var trackerEventStrings []string
	for _, trackerEvent := range replayData.TrackerEvts.Evts {
		trackerEventStrings = append(trackerEventStrings, trackerEvent.String())
	}

	// These structures are handled differently as it is a Map without .String() method:
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
		PIDPlayerDescMapStrings = append(PIDPlayerDescMapStrings, "\""+playerNumber+"\": "+string(playerDescInformation))
	}

	var ToonPlayerDescMapStrings []string
	for ToonPlayerDescKey, ToonPlayerDescValue := range ToonPlayerDescMap {

		// Converting ID to string:
		playerToon := ToonPlayerDescKey

		// Converting struct to JSON:
		playerDescInformation, err := json.Marshal(ToonPlayerDescValue)

		if err != nil {
			// Return false
			panic(err)
		}

		// Putting everything together:
		ToonPlayerDescMapStrings = append(ToonPlayerDescMapStrings, "\""+playerToon+"\": "+string(playerDescInformation))
	}

	// Booleans saying if processing had any errors
	gameEvtsErr := strconv.FormatBool(replayData.GameEvtsErr)
	messageEvtsErr := strconv.FormatBool(replayData.MessageEvtsErr)
	trackerEvtsErr := strconv.FormatBool(replayData.TrackerEvtsErr)

	// Crezting JSON structure by hand:
	var strBuilder strings.Builder
	fmt.Fprintf(&strBuilder, "{\n")
	fmt.Fprintf(&strBuilder, "  \"header\": %s,\n", header)
	fmt.Fprintf(&strBuilder, "  \"initData\": %s,\n", initData)
	fmt.Fprintf(&strBuilder, "  \"details\": %s,\n", details)
	fmt.Fprintf(&strBuilder, "  \"attrEvts\": %s,\n", attrEvts)
	fmt.Fprintf(&strBuilder, "  \"metadata\": %s,\n", metadata)
	fmt.Fprintf(&strBuilder, "  \"gameEvtsErr\": %s,\n", gameEvtsErr)
	fmt.Fprintf(&strBuilder, "  \"messageEvtsErr\": %s,\n", messageEvtsErr)
	fmt.Fprintf(&strBuilder, "  \"trackerEvtsErr\": %s,\n", trackerEvtsErr)
	fmt.Fprintf(&strBuilder, "  \"messageEventsStrings\": [%s],\n", strings.Join(messageEventStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"gameEventStrings\": [%s],\n", strings.Join(gameEventStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"trackerEventStrings\": [%s],\n", strings.Join(trackerEventStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"PIDPlayerDescMap\": {%s},\n", strings.Join(PIDPlayerDescMapStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"ToonPlayerDescMap\": {%s},\n", strings.Join(ToonPlayerDescMapStrings, ",\n"))
	fmt.Fprintf(&strBuilder, "  \"gameEvtsErr\": %s", gameEvtsErr+",\n")
	fmt.Fprintf(&strBuilder, "  \"messageEvtsErr\": %s", messageEvtsErr+",\n")
	fmt.Fprintf(&strBuilder, "  \"trackerEvtsErr\": %s", trackerEvtsErr+"\n")
	fmt.Fprintf(&strBuilder, "  ")
	fmt.Fprintf(&strBuilder, "}")

	// TODO: Return a summary in a custom Golang struct.
	return true, strBuilder.String()
}
