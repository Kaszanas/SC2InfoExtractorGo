package dataproc

import (
	"context"
	"fmt"
	"time"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	pb "github.com/Kaszanas/GoSC2Science/proto"
	settings "github.com/Kaszanas/GoSC2Science/settings"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func anonymizeReplay(replayData *data.CleanedReplay, playersAnonymized *map[string]int) bool {

	log.Info("Entered anonymizeReplay()")

	if !anonimizeMessageEvents(replayData) {
		log.Error("Failed to anonimize messageEvents.")
		return false
	}

	if !anonymizePlayers(replayData, playersAnonymized) {
		log.Error("Failed to anonimize player information.")
		return false
	}

	log.Info("Finished anonymizeReplay()")
	return true
}

// grpcConnectAnonymize is using https://github.com/Kaszanas/SC2AnonServerPy in order to anonymize users.
func grpcConnectAnonymize(toonString string) string {
	// Set up a connection to the server:
	conn, err := grpc.Dial(settings.GrpcServerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Start-up a gRPC client:
	c := pb.NewAnonymizeServiceClient(conn)

	// Contact the server and print out its response:
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := c.GetAnonymizedID(ctx, &pb.SendNickname{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.WithField("gRPC_response", result.AnonymizedID).Debug("Received anonymized ID for a player.")
	return result.AnonymizedID
}

func anonymizePlayers(replayData *data.CleanedReplay) bool {

	log.Info("Entererd anonymizePlayers().")

	var newToonDescMap = make(map[string]*rep.PlayerDesc)
	// Connecting to gRPC server:

	// Iterate over players:
	log.Info("Starting to iterate over replayData.Details.PlayerList.")

	// TODO: Iterating over PlayerList might not be the best IDEA!!!!!!!
	// There is absolutely no assurance when it comes to the ordering of the players.
	for _, playerData := range replayData.Metadata.Players {
		// Iterate over Toon description map:
		for toon, playerDesc := range replayData.ToonPlayerDescMap {
			// Checking if the SlotID and TeamID matches:
			if playerDesc.PlayerID == int64(playerData.PlayerID) {
				grpcConnectAnonymize(toon)
				// Checking if the player toon was already anonymized (toons are unique, nicknames are not)
				// TODO: Use gRPC here!
			}
		}
	}

	// Replacing Toon desc map with anonymmized version containing a persistent anonymized ID of the player:
	log.Info("Replacing ToonPlayerDescMap with anonymized version.")
	replayData.ToonPlayerDescMap = newToonDescMap

	fmt.Println(replayData.ToonPlayerDescMap)

	log.Info("Finished anonymizePlayers()")
	return true
}

func anonimizeMessageEvents(replayData *data.CleanedReplay) bool {

	log.Info("Entered anonimizeMessageEvents().")
	var anonymizedMessageEvents []s2prot.Struct
	for _, event := range replayData.MessageEvents {
		eventType := event["evtTypeName"].(string)
		if !contains(settings.UnusedMessageEvents, eventType) {
			anonymizedMessageEvents = append(anonymizedMessageEvents, event)
		}
	}

	replayData.MessageEvents = anonymizedMessageEvents

	log.Info("Finished anonymizeMessageEvents()")
	return true
}

func anonymizeToonDescMap(playerDesc *rep.PlayerDesc, toonDescMap *map[string]*rep.PlayerDesc, anonymizedID string) {

	log.Info("Entered anonymizeToonDescMap().")

	// Define new rep.PlayerDesc with old
	emptyPlayerDesc := rep.PlayerDesc{
		PlayerID:            playerDesc.PlayerID,
		SlotID:              playerDesc.SlotID,
		UserID:              playerDesc.UserID,
		StartLocX:           playerDesc.StartLocX,
		StartLocY:           playerDesc.StartLocY,
		StartDir:            playerDesc.StartDir,
		SQ:                  playerDesc.SQ,
		SupplyCappedPercent: playerDesc.SupplyCappedPercent,
	}

	// Adding the new PlayerDesc
	log.Info("Adding new PlayerDesc to toonDescMap")
	(*toonDescMap)[anonymizedID] = &emptyPlayerDesc

	log.Info("Finished anonymizeToonDescMap()")
}
