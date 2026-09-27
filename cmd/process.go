package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// processingOptions holds the flags that control how replays are processed,
// shared by `process` and `process_replay`.
type processingOptions struct {
	skipDependencyDownload     bool
	performIntegrityCheck      bool
	performValidityCheck       bool
	performCleanup             bool
	performPlayerAnonymization bool
	performChatAnonymization   bool
	performFiltering           bool
	gameModeFilter             int
}

// addProcessingFlags registers the processing flags on the given flag set.
func addProcessingFlags(flags *pflag.FlagSet, opts *processingOptions) {
	flags.BoolVar(&opts.skipDependencyDownload, "skip_dependency_download", false,
		"Skip downloading the replay dependencies, use only what is already in --dependency_directory.")
	flags.BoolVar(&opts.performIntegrityCheck, "perform_integrity_checks", false,
		"Run the hardcoded integrity checks on the replays.")
	flags.BoolVar(&opts.performValidityCheck, "perform_validity_checks", false,
		"Check that replay values are within 'common sense' ranges.")
	flags.BoolVar(&opts.performCleanup, "perform_cleanup", false,
		"Run the cleaning functions of the processing pipeline.")
	flags.BoolVar(&opts.performPlayerAnonymization, "perform_player_anonymization", false,
		"Anonymize players. Requires a running anonymization server: https://doi.org/10.5281/zenodo.5138313")
	flags.BoolVar(&opts.performChatAnonymization, "perform_chat_anonymization", false,
		"Anonymize chat messages.")
	flags.BoolVar(&opts.performFiltering, "perform_filtering", false,
		"Filter replays by game mode (see --game_mode_filter). If not set, no filtering is performed.")
	flags.IntVar(&opts.gameModeFilter, "game_mode_filter", 0b11111111,
		"Game modes to include, as a binary flag. All game modes: 0b11111111.")
}

// apply copies the processing options into the pipeline configuration.
func (o *processingOptions) apply(flags *utils.CLIFlags) {
	flags.SkipDependencyDownload = o.skipDependencyDownload
	flags.PerformIntegrityCheck = o.performIntegrityCheck
	flags.PerformValidityCheck = o.performValidityCheck
	flags.PerformCleanup = o.performCleanup
	flags.PerformPlayerAnonymization = o.performPlayerAnonymization
	flags.PerformChatAnonymization = o.performChatAnonymization
	flags.PerformFiltering = o.performFiltering
	flags.FilterGameMode = o.gameModeFilter
}

// processOptions holds the flags shared by every output format of `process`.
type processOptions struct {
	root            *rootOptions
	outputDirectory string
	processing      processingOptions
}

func newProcessCmd(rootOpts *rootOptions, runner Runner) *cobra.Command {
	opts := &processOptions{root: rootOpts}

	processCmd := &cobra.Command{
		Use:   "process",
		Short: "Process replays into one of the output formats.",
		Long: "Process replays into one of the output formats. " +
			"Each output format is its own subcommand.",
		RunE: requireSubcommand,
	}

	flags := processCmd.PersistentFlags()
	flags.StringVar(&opts.outputDirectory, "output", "./replays/output",
		"Output directory for the processed replays.")
	addProcessingFlags(flags, &opts.processing)

	processCmd.AddCommand(
		newProcessJSONCmd(opts, runner),
		newProcessJSONZipCmd(opts, runner),
		newProcessSingleJSONCmd(opts, runner),
	)
	return processCmd
}

// flags builds the pipeline configuration shared by every output format.
func (o *processOptions) flags() (utils.CLIFlags, error) {
	flags, err := o.root.baseFlags()
	if err != nil {
		return utils.CLIFlags{}, err
	}

	outputDirectory, err := filepath.Abs(o.outputDirectory)
	if err != nil {
		return utils.CLIFlags{}, fmt.Errorf("resolving --output %q: %w", o.outputDirectory, err)
	}

	flags.OutputDirectory = outputDirectory
	o.processing.apply(&flags)
	return flags, nil
}
