package cmd

import "github.com/spf13/cobra"

func newProcessJSONCmd(opts *processOptions, runner Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "json",
		Short: "Write one .json file per replay.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags, err := opts.flags()
			if err != nil {
				return err
			}
			flags.NumberOfPackages = 0
			return runner.Process(flags)
		},
	}
}
