package dataproc

import (
	"context"
	"time"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	pb "github.com/Kaszanas/SC2InfoExtractorGo/proto"
	settings "github.com/Kaszanas/SC2InfoExtractorGo/settings"
	"github.com/icza/s2prot"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// checkAnonymizationInitializeGRPC verifies if the anonymization should
// be performed and returns a pointer to GRPCAnonymizer.
func checkAnonymizationInitializeGRPC(
	performAnonymizationBool bool,
) *GRPCAnonymizer {
	if !performAnonymizationBool {
		return nil
	}

	log.Info("Detected that user wants anonymization, attempting to set up GRPCAnonymizer{}")
	grpcAnonymizer := GRPCAnonymizer{}
	if !grpcAnonymizer.grpcDialConnect() {
		log.Error("Could not connect to the gRPC server!")
	}
	grpcAnonymizer.grpcInitializeClient()
	grpcAnonymizer.Cache = make(map[string]string)

	return &grpcAnonymizer
}

// anonymizeReplay is the main function that is responsible for
// anonymizing the replay data. It calls other functions that are
// responsible for anonymizing chat messages and player information.
func anonymizeReplay(
	replayData *replay_data.CleanedReplay,
	grpcAnonymizer *GRPCAnonymizer,
	performChatAnonymizationBool bool,
	performPlayerAnonymizationBool bool,
) bool {

	log.Debug("Entered anonymizeReplay()")

	// Anonymization of Chat events that might
	// contain sensitive information for research purposes:
	if performChatAnonymizationBool {
		if !anonimizeMessageEvents(replayData) {
			log.Error("Failed to anonimize messageEvents.")
			return false
		}
	}

	// Anonymizing player information such as toon, nickname,
	// and clan this is done in order to redact potentially sensitive information:
	if performPlayerAnonymizationBool {
		if !anonymizePlayers(replayData, grpcAnonymizer) {
			log.Error("Failed to anonimize player information.")
			return false
		}
	}

	log.Debug("Finished anonymizeReplay()")
	return true
}

//nolint:all
var keepAliveParameters = keepalive.ClientParameters{
	Time:                20 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             10 * time.Second, // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// Create new class, AnonymizerClient, that wraps the gRPC client
// (pb.NewAnonymizeServiceClient(conn) should happen once).
// The class will store the gRPC connection and can store a local cache of responses.
type GRPCAnonymizer struct {
	Connection *grpc.ClientConn
	Client     pb.AnonymizeServiceClient
	Cache      map[string]string
}

// grpcConnect initializes a connection to a specified in settings grpc server.
func (anonymizer *GRPCAnonymizer) grpcDialConnect() bool {

	log.Debug("Entered GRPCAnonymizer.grpcDialConnect()")

	// Set up a connection to the server:
	conn, err := grpc.NewClient(settings.GrpcServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.WithField("error", err).
			Fatal("Failed to connect to grpc anonymization service")
		return false
	}

	anonymizer.Connection = conn

	log.Debug("Finished GRPCAnonymizer.grpcDialConnect()")
	return true

}

// grpcInitializeClient initializes a client for the gRPC connection.
func (anonymizer *GRPCAnonymizer) grpcInitializeClient() {
	// Initialize a grpcClient
	anonymizer.Client = pb.NewAnonymizeServiceClient(anonymizer.Connection)
}

// anonymizeToon checks if the player toon is already
// in the cache and if it is not it calls grpcAnonymizeID.
func (anonymizer *GRPCAnonymizer) anonymizeToon(toonString string) (string, bool) {

	log.Debug("Entered GRPCAnonymizer.anonymizeToon()")

	// Check if the toon is already in cache not to spam the connection with requests:
	val, ok := anonymizer.Cache[toonString]
	if ok {
		return val, true
	}

	// If the toonString is not within cache already we check
	// if it is possible to obtain it from anonymization server:
	anonymizedID, grpcAnonOk := grpcGetAnonymizeID(
		toonString,
		anonymizer.Client,
		anonymizer.Connection)
	if !grpcAnonOk {
		return "", false
	}
	anonymizer.Cache[toonString] = anonymizedID
	log.Debug("Finished GRPCAnonymizer.anonymizeToon()")

	return anonymizedID, true
}

// grpcGetAnonymizeID is using https://github.com/Kaszanas/SC2AnonServerPy
// in order to anonymize users.
func grpcGetAnonymizeID(
	toonString string,
	grpcClient pb.AnonymizeServiceClient,
	grpcConnection *grpc.ClientConn,
) (string, bool) {

	log.Debug("Entered grpcAnonymize()")

	// Contact the server and print out its response:
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	result, err := grpcClient.GetAnonymizedID(
		ctx,
		&pb.SendNickname{Nickname: toonString})
	if err != nil {
		log.WithField("error", err).
			Fatalf("Could not receive anonymized information from grpc service!")
		return "", false
	}
	log.WithField("gRPC_response", result.AnonymizedID).
		Debug("Received anonymized ID for a player.")

	log.Debug("Finished grpcAnonymize()")
	return result.AnonymizedID, true
}

func anonymizePlayers(
	replayData *replay_data.CleanedReplay,
	grpcAnonymizer *GRPCAnonymizer) bool {

	log.Info("Entererd anonymizePlayers().")

	var newToonDescMap = make(map[string]replay_data.EnhancedToonDescMap)

	// Iterate over players:
	log.Info("Starting to iterate over replayData.Details.PlayerList.")

	// Iterate over Toon description map:
	for toon, playerDesc := range replayData.ToonPlayerDescMap {

		// Using gRPC for anonymization:
		anonymizedID, anonymizeToonOk := grpcAnonymizer.anonymizeToon(toon)
		if !anonymizeToonOk {
			return false
		}
		anonymizedPlayerDesc := playerDesc
		anonymizedPlayerDesc.Name = "redacted"
		anonymizedPlayerDesc.ClanTag = "redacted"

		newToonDescMap[anonymizedID] = anonymizedPlayerDesc

	}

	// Replacing Toon desc map with anonymmized version containing
	// a persistent anonymized ID of the player:
	log.Info("Replacing ToonPlayerDescMap with anonymized version.")
	replayData.ToonPlayerDescMap = newToonDescMap

	log.WithField("toonDescMapAnonymized", replayData.ToonPlayerDescMap).
		Debug("Replaced toonDescMap with anonymized version")

	log.Debug("Finished anonymizePlayers()")
	return true
}

// anonymizeMessageEvents checks against settings.UnusedMessageEvents
// and creates a new clean version without specified events.
func anonimizeMessageEvents(replayData *replay_data.CleanedReplay) bool {

	log.Debug("Entered anonimizeMessageEvents().")
	var anonymizedMessageEvents []s2prot.Struct
	for _, event := range replayData.MessageEvents {
		eventType := event["evtTypeName"].(string)
		if !contains(settings.AnonymizeMessageEvents, eventType) {
			anonymizedMessageEvents = append(anonymizedMessageEvents, event)
		}
	}

	replayData.MessageEvents = anonymizedMessageEvents

	log.Debug("Finished anonymizeMessageEvents()")
	return true
}
