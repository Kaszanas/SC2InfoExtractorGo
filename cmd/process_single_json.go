package cmd

import "github.com/spf13/cobra"

func newProcessSingleJSONCmd(opts *processOptions, runner Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "single_json",
		Short: "Write all replays into a single all_replays.json array.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags, err := opts.flags()
			if err != nil {
				return err
			}
			flags.SingleJsonOutput = true
			// A single chunk keeps the order of replays in the output
			// deterministic (the order of the input files):
			flags.NumberOfPackages = 1
			return runner.Process(flags)
		},
	}
}
