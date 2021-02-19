package dataproc

import "github.com/icza/s2prot/rep"

func generateSummary(replayData *rep.Rep) {
	// TODO: Prepare a summary for the replay that was processed

	// Game version histogram (This needs to be created on a file by file basis)

	// Game time histogram (This should take game duration into consideration in seconds or possibly every 5 seconds to decrease the number of datapoints)

	// Maps used histogram (This needs to take into consideration that the maps might be named differently depending on what language version of the game was used?)
	// This might require using the map checksums or some other additional information to synchronize.

	// Race summary (This will be calculated on a replay by replay basis)

	// Amount of different units used (histogram of units used). Is this needed?

	// Dates of the replay when was the first recorded replay in the package when was the last recorded replay in the package.

	// Server information histogram. Region etc.

	// Histograms for maximum game time in different matchups. PvP, ZvP, TvP, ZvT, TvT, ZvZ

	// How many unique accounts were found

}
