// Package cmd defines the command line interface of SC2InfoExtractorGo.
package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// rootOptions holds the flags shared by every command.
type rootOptions struct {
	inputDirectory      string
	dependencyDirectory string
	maxProcs            int
	logDirectory        string
	logLevel            int
	cpuProfilingPath    string
}

// NewRootCmd builds the full command tree. The runner performs the actual
// work, which lets tests verify the flag handling without processing replays.
func NewRootCmd(runner Runner) *cobra.Command {
	opts := &rootOptions{}

	rootCmd := &cobra.Command{
		Use:          "SC2InfoExtractorGo",
		Short:        "Extracts data from StarCraft II .SC2Replay files.",
		SilenceUsage: true,
		Args:         unknownCommand,
		RunE:         requireSubcommand,
	}
	rootCmd.SetGlobalNormalizationFunc(snakeCaseFlags)
	rootCmd.SetFlagErrorFunc(flagErrorWithHint)

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&opts.inputDirectory, "input", "./replays/input",
		"Input directory where .SC2Replay files are held.")
	flags.StringVar(&opts.dependencyDirectory, "dependency_directory", "./dependencies/",
		"Directory where the replay dependencies are downloaded to and read from.")
	flags.IntVar(&opts.maxProcs, "max_procs", runtime.NumCPU(),
		"Number of logical processor cores used for processing.")
	flags.StringVar(&opts.logDirectory, "log_dir", "./logs/",
		"Directory which will hold the logging information.")
	flags.IntVar(&opts.logLevel, "log_level", 3,
		"Log level from 0-6: Panic - 0, Fatal - 1, Error - 2, Warn - 3, Info - 4, Debug - 5, Trace - 6.")
	flags.StringVar(&opts.cpuProfilingPath, "with_cpu_profiler", "",
		"Path to the file where the pprof CPU profile will be saved. If empty, no profiling is performed.")

	rootCmd.AddCommand(
		newDownloadDepsCmd(opts, runner),
		newProcessCmd(opts, runner),
		newProcessReplayCmd(opts, runner),
	)
	return rootCmd
}

// baseFlags converts the shared flags into the configuration used by the pipeline.
func (o *rootOptions) baseFlags() (utils.CLIFlags, error) {
	inputDirectory, err := filepath.Abs(o.inputDirectory)
	if err != nil {
		return utils.CLIFlags{}, fmt.Errorf("resolving --input %q: %w", o.inputDirectory, err)
	}
	dependencyDirectory, err := filepath.Abs(o.dependencyDirectory)
	if err != nil {
		return utils.CLIFlags{}, fmt.Errorf(
			"resolving --dependency_directory %q: %w", o.dependencyDirectory, err)
	}

	return utils.CLIFlags{
		InputDirectory:      inputDirectory,
		DependencyDirectory: dependencyDirectory,
		NumberOfThreads:     o.maxProcs,
		LogFlags: utils.LogFlags{
			LogLevelValue: datastruct.LogLevelEnum(o.logLevel),
			LogPath:       o.logDirectory,
		},
		CPUProfilingPath: o.cpuProfilingPath,
	}, nil
}

// requireSubcommand is used by commands that only group other commands.
// Called bare they print their help; called with arguments or flags they fail,
// so that old flag-only invocations don't silently do nothing.
func requireSubcommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && cmd.Flags().NFlag() == 0 {
		return cmd.Help()
	}
	return fmt.Errorf("%q needs a subcommand, see %q",
		cmd.CommandPath(), cmd.CommandPath()+" --help")
}

// unknownCommand rejects positional arguments to the root command. Setting
// Args explicitly makes cobra parse flags before validating arguments, so the
// old single-dash syntax (-input ./x) fails as a flag error with a hint rather
// than as "unknown command ./x".
func unknownCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	message := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())
	if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		message += fmt.Sprintf(", did you mean %q?", suggestions[0])
	}
	return fmt.Errorf("%s\nsee %q", message, cmd.CommandPath()+" --help")
}

// snakeCaseFlags lets kebab-case spellings (--log-level) resolve to the
// documented snake_case flags (--log_level).
func snakeCaseFlags(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.ReplaceAll(name, "-", "_"))
}

// flagErrorWithHint points users of the old single-dash, command-less syntax
// to the current one.
func flagErrorWithHint(cmd *cobra.Command, err error) error {
	return fmt.Errorf(
		"%w\nflags take two dashes and follow a command, e.g. "+
			"\"SC2InfoExtractorGo process json_zip --input ./replays/input\"; see %q",
		err, cmd.CommandPath()+" --help")
}
