package cleanup

import (
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// cleanMessageEvents copies the message events,
// has the capability of removing unescessary fields.
func CleanMessageEvents(replayData *rep.Rep) []s2prot.Struct {
	// Constructing a clean MessageEvents without unescessary fields:
	var messageEventsStructs []s2prot.Struct
	for _, messageEvent := range replayData.MessageEvts {
		messageEventsStructs = append(messageEventsStructs, messageEvent.Struct)
	}
	log.Info("Defined cleanMessageEvents struct")
	return messageEventsStructs
}
