package cmd

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc"
	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/chunk_utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

// compressionMethod is the zip compression method used for output packages.
const compressionMethod = zip.Deflate

// Runner performs the work behind the commands.
type Runner interface {
	// Process downloads the replay dependencies and processes the replays.
	Process(flags utils.CLIFlags) error
	// DownloadDeps only downloads the replay dependencies.
	DownloadDeps(flags utils.CLIFlags) error
	// ProcessReplay processes a single replay file and returns its JSON.
	ProcessReplay(flags utils.CLIFlags, replayFile string) (string, error)
}

type defaultRunner struct{}

// DefaultRunner returns the Runner that runs the real processing pipeline.
func DefaultRunner() Runner {
	return defaultRunner{}
}

func (defaultRunner) DownloadDeps(flags utils.CLIFlags) error {
	return withRuntimeSetup(flags, func() error {
		listOfInputFiles, err := listReplayFiles(flags)
		if err != nil {
			return err
		}

		downloader.DependencyDownloaderPipeline(
			listOfInputFiles,
			foreignToEnglishMappingPath(flags),
			flags,
		)
		log.Info("Only dependency download was chosen. Exiting.")
		return nil
	})
}

func (defaultRunner) Process(flags utils.CLIFlags) error {
	return withRuntimeSetup(flags, func() error {
		listOfInputFiles, err := listReplayFiles(flags)
		if err != nil {
			return err
		}

		lenListOfInputFiles := len(listOfInputFiles)
		if lenListOfInputFiles < flags.NumberOfPackages {
			return fmt.Errorf(
				"--number_of_packages (%d) is higher than the number of input files (%d)",
				flags.NumberOfPackages, lenListOfInputFiles)
		}

		foreignToEnglishMapping := downloader.DependencyDownloaderPipeline(
			listOfInputFiles,
			foreignToEnglishMappingPath(flags),
			flags,
		)

		listOfChunksFiles, packageToZipBool := chunk_utils.GetChunkListAndPackageBool(
			listOfInputFiles,
			flags.NumberOfPackages,
			flags.NumberOfThreads,
			lenListOfInputFiles,
		)

		dataproc.PipelineWrapper(
			listOfChunksFiles,
			packageToZipBool,
			compressionMethod,
			foreignToEnglishMapping,
			flags,
		)
		return nil
	})
}

func (defaultRunner) ProcessReplay(flags utils.CLIFlags, replayFile string) (string, error) {
	var replayJSON string
	err := withRuntimeSetup(flags, func() error {
		foreignToEnglishMapping := downloader.DependencyDownloaderPipeline(
			[]string{replayFile},
			foreignToEnglishMappingPath(flags),
			flags,
		)

		var err error
		replayJSON, err = dataproc.ProcessReplayToJSON(
			replayFile,
			foreignToEnglishMapping,
			flags,
		)
		return err
	})
	return replayJSON, err
}

// withRuntimeSetup sets up file logging and optional CPU profiling around run,
// and tears both down afterwards.
func withRuntimeSetup(flags utils.CLIFlags, run func() error) error {
	// Status goes to stderr so stdout carries only command output
	// (process_replay prints the replay JSON there):
	fmt.Fprintln(os.Stderr, "SC2InfoExtractorGo started.")

	logFile, okLogging := utils.SetLogging(
		flags.LogFlags.LogPath,
		int(flags.LogFlags.LogLevelValue),
	)
	if !okLogging {
		return errors.New("failed to set up logging")
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			log.WithField("error", err).Error("Failed to close log file.")
		}
	}()

	log.WithField("flags", fmt.Sprintf("%+v", flags)).Info("Parsed command line flags")

	// Profiling capabilities to verify if the program can be optimized any further:
	if flags.CPUProfilingPath != "" {
		_, okProfiling := utils.SetProfiling(flags.CPUProfilingPath)
		if !okProfiling {
			return errors.New("failed to set up CPU profiling")
		}
		defer pprof.StopCPUProfile()
	}

	return run()
}

// listReplayFiles lists the .SC2Replay files in the input directory.
func listReplayFiles(flags utils.CLIFlags) ([]string, error) {
	listOfInputFiles, err := file_utils.ListFiles(flags.InputDirectory, ".SC2Replay")
	if err != nil {
		return nil, fmt.Errorf("listing replays in %q: %w", flags.InputDirectory, err)
	}
	return listOfInputFiles, nil
}

// foreignToEnglishMappingPath places the map name mapping next to the log file.
func foreignToEnglishMappingPath(flags utils.CLIFlags) string {
	return flags.LogFlags.LogPath + "map_foreign_to_english_mapping.json"
}
