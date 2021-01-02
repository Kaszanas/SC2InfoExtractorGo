package main

import (
	"fmt"
	// "github.com/icza/mpq"
	"encoding/json"
	// "github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	// "log"
)

func main() {

	// Getting list of files from input directory:
	listOfInputFiles := listFiles("./DEMOS/Input")

	for _, file := range listOfInputFiles {
		fmt.Println(file.Name())
		absoluteReplayFilepath, err := filepath.Abs(file.Name())
		if err != nil {
			fmt.Printf("Failed to provide absolute_path: %v\n", err)
		} else {
			// fmt.Println(absoluteReplayFilepath)
			// replayFilepath := "./DEMOS/Input/11506446_1566325366_8429955.SC2Replay"

			replayFile, err := rep.NewFromFile(absoluteReplayFilepath)
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
			ToonPlayerDescMap := replayFile.TrackerEvts.ToonPlayerDescMap

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
					panic(err)
				}

				// Putting everything together:
				ToonPlayerDescMapStrings = append(ToonPlayerDescMapStrings, "\""+playerToon+"\": "+string(playerDescInformation))
			}

			// Booleans saying if processing had any errors
			gameEvtsErr := strconv.FormatBool(replayFile.GameEvtsErr)
			messageEvtsErr := strconv.FormatBool(replayFile.MessageEvtsErr)
			trackerEvtsErr := strconv.FormatBool(replayFile.TrackerEvtsErr)

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

			// Writing JSON file:
			_ = ioutil.WriteFile("./DEMOS/Output/11506446_1566325366_8429955.json", []byte(strBuilder.String()), 0644)

			// TODO: every event

		}
	}
}
