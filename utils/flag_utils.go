package utils

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
)

// LogFlags contains settings that the user can set for logging.
type LogFlags struct {
	LogLevelValue datastruct.LogLevelEnum
	LogPath       string
}

// CLIFlags is a structure which holds all of the information that was supplied by user in CLI.
// It is populated by the commands in the cmd package.
type CLIFlags struct {
	InputDirectory             string
	OutputDirectory            string
	OnlyDependencyDownload     bool
	SkipDependencyDownload     bool
	DependencyDirectory        string
	NumberOfThreads            int
	NumberOfPackages           int
	SingleJsonOutput           bool
	PerformIntegrityCheck      bool
	PerformValidityCheck       bool
	PerformCleanup             bool
	PerformPlayerAnonymization bool
	PerformChatAnonymization   bool
	PerformFiltering           bool
	FilterGameMode             int
	LogFlags                   LogFlags
	CPUProfilingPath           string
}
