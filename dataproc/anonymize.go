package dataproc

import (
	"context"
	"time"

	data "github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	pb "github.com/Kaszanas/SC2InfoExtractorGo/proto"
	settings "github.com/Kaszanas/SC2InfoExtractorGo/settings"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func anonymizeReplay(replayData *data.CleanedReplay, grpcAnonymizer *GRPCAnonymizer) bool {

	log.Info("Entered anonymizeReplay()")

	// Anonymization of Chat events that might contain sensitive information for research purposes:
	if !anonimizeMessageEvents(replayData) {
		log.Error("Failed to anonimize messageEvents.")
		return false
	}

	// Anonymizing player information such as toon, nickname, and clan this is done in order to redact potentially sensitive information:
	if !anonymizePlayers(replayData, grpcAnonymizer) {
		log.Error("Failed to anonimize player information.")
		return false
	}

	log.Info("Finished anonymizeReplay()")
	return true
}

var keepAliveParameters = keepalive.ClientParameters{
	Time:                5 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             2 * time.Second, // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,            // send pings even without active streams
}

// Create new class, AnonymizerClient, that wraps the gRPC client (pb.NewAnonymizeServiceClient(conn) should happen once).
// The class will store the gRPC connection and can store a local cache of responses.
type GRPCAnonymizer struct {
	Connection *grpc.ClientConn
	Client     pb.AnonymizeServiceClient
	Cache      map[string]string
}

// grpcConnect initializes a connection to a specified in settings grpc server.
func (anonymizer *GRPCAnonymizer) grpcDialConnect() bool {

	log.Info("Entered GRPCAnonymizer.grpcDialConnect()")

	// Set up a connection to the server:
	conn, err := grpc.Dial(settings.GrpcServerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.WithField("error", err).Fatal("Failed to connect to grpc anonymization service")
		return false
	}

	anonymizer.Connection = conn

	log.Info("Finished GRPCAnonymizer.grpcDialConnect()")
	return true

}

func (anonymizer *GRPCAnonymizer) grpcInitializeClient() {
	// Initialize a grpcClient
	anonymizer.Client = pb.NewAnonymizeServiceClient(anonymizer.Connection)
}

// anonymizeToon checks if the player toon is already in the cache and if it is not it calls grpcAnonymizeID.
func (anonymizer *GRPCAnonymizer) anonymizeToon(toonString string) (string, bool) {

	log.Info("Entered GRPCAnonymizer.anonymizeToon()")

	// Check if the toon is already in cache not to spam the connection with requests:
	val, ok := anonymizer.Cache[toonString]
	if ok {
		return val, true
	}

	// If the toonString is not within cache already we check if it is possible to obtain it from anonymization server:
	anonymizedID, grpcAnonOk := grpcGetAnonymizeID(toonString, anonymizer.Client, anonymizer.Connection)
	if !grpcAnonOk {
		return "", false
	}
	anonymizer.Cache[toonString] = anonymizedID
	log.Info("Finished GRPCAnonymizer.anonymizeToon()")

	return anonymizedID, true
}

// grpcGetAnonymizeID is using https://github.com/Kaszanas/SC2AnonServerPy in order to anonymize users.
func grpcGetAnonymizeID(toonString string, grpcClient pb.AnonymizeServiceClient, grpcConnection *grpc.ClientConn) (string, bool) {

	log.Info("Entered grpcAnonymize()")

	// Contact the server and print out its response:
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	result, err := grpcClient.GetAnonymizedID(ctx, &pb.SendNickname{Nickname: toonString})
	if err != nil {
		log.WithField("error", err).Fatalf("Could not receive anonymized information from grpc service!")
		return "", false
	}
	log.WithField("gRPC_response", result.AnonymizedID).Debug("Received anonymized ID for a player.")

	log.Info("Finished grpcAnonymize()")
	return result.AnonymizedID, true
}

func anonymizePlayers(replayData *data.CleanedReplay, grpcAnonymizer *GRPCAnonymizer) bool {

	log.Info("Entererd anonymizePlayers().")

	var newToonDescMap = make(map[string]data.EnhancedToonDescMap)

	// Iterate over players:
	log.Info("Starting to iterate over replayData.Details.PlayerList.")

	// Iterate over Toon description map:
	for toon, playerDesc := range replayData.ToonPlayerDescMap {

		// Using gRPC for anonymization:
		anonymizedID := grpcAnonymizer.anonymizeToon(toon)
		anonymizedPlayerDesc := playerDesc
		anonymizedPlayerDesc.Name = "redacted"
		anonymizedPlayerDesc.ClanTag = "redacted"

		newToonDescMap[anonymizedID] = anonymizedPlayerDesc

	}

	// Replacing Toon desc map with anonymmized version containing a persistent anonymized ID of the player:
	log.Info("Replacing ToonPlayerDescMap with anonymized version.")
	replayData.ToonPlayerDescMap = newToonDescMap

	log.WithField("toonDescMapAnonymized", replayData.ToonPlayerDescMap).Debug("Replaced toonDescMap with anonymized version")

	// fmt.Println(replayData.ToonPlayerDescMap)

	log.Info("Finished anonymizePlayers()")
	return true
}

// anonymizeMessageEvents checks against settings.UnusedMessageEvents and creates a new clean version without specified events.
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

// TODO: This could be deleted?
// anonymizeToonDescMap is a deprecated version that was used for a single threaded approach in ToonDescMap anonymization.
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
