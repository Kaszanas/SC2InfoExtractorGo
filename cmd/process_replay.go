package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// stdoutOutput is the --output value that selects stdout.
const stdoutOutput = "-"

func newProcessReplayCmd(rootOpts *rootOptions, runner Runner) *cobra.Command {
	var output string
	var processing processingOptions

	cmd := &cobra.Command{
		Use:   "process_replay <file.SC2Replay>",
		Short: "Process a single replay into JSON, printed to stdout or written to --output.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("input") {
				return errors.New(
					"--input is not used by process_replay, pass the replay file as an argument")
			}

			replayFile, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("resolving replay path %q: %w", args[0], err)
			}
			info, err := os.Stat(replayFile)
			if err != nil {
				return fmt.Errorf("reading replay: %w", err)
			}
			if info.IsDir() {
				return fmt.Errorf(
					"%q is a directory, use \"process\" to process a directory of replays", args[0])
			}

			flags, err := rootOpts.baseFlags()
			if err != nil {
				return err
			}
			flags.InputDirectory = filepath.Dir(replayFile)
			processing.apply(&flags)

			replayJSON, err := runner.ProcessReplay(flags, replayFile)
			if err != nil {
				return err
			}
			return writeReplayJSON(cmd.OutOrStdout(), output, replayFile, replayJSON)
		},
	}

	cmd.Flags().StringVar(&output, "output", stdoutOutput,
		"Where to write the JSON: a file path, an existing directory "+
			"(writes <dir>/<replay name>.json), or - for stdout.")
	addProcessingFlags(cmd.Flags(), &processing)
	return cmd
}

// writeReplayJSON writes the replay JSON to stdout or to the file or
// directory given by output.
func writeReplayJSON(stdout io.Writer, output, replayFile, replayJSON string) error {
	if output == stdoutOutput || output == "" {
		_, err := fmt.Fprintln(stdout, replayJSON)
		return err
	}

	outputPath := output
	if info, err := os.Stat(output); err == nil && info.IsDir() {
		// Same naming as the per-replay files of `process json`:
		replayName := strings.TrimSuffix(filepath.Base(replayFile), filepath.Ext(replayFile))
		outputPath = filepath.Join(output, replayName+".json")
	} else if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("creating the output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(replayJSON), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outputPath, err)
	}
	return nil
}
