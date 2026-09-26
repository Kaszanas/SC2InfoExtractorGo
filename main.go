package main

import (
	"os"

	"github.com/Kaszanas/SC2InfoExtractorGo/cmd"
)

func main() {
	// Cobra reports command line errors itself; file-based logging is set up
	// by the commands that process replays.
	if err := cmd.NewRootCmd(cmd.DefaultRunner()).Execute(); err != nil {
		os.Exit(1)
	}
}
