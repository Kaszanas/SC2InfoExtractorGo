package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newProcessJSONZipCmd(opts *processOptions, runner Runner) *cobra.Command {
	var numberOfPackages int

	cmd := &cobra.Command{
		Use:   "json_zip",
		Short: "Write replays as .json files inside zip packages, with a summary per package.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if numberOfPackages < 1 {
				return fmt.Errorf(
					"--number_of_packages must be at least 1, got %d", numberOfPackages)
			}
			flags, err := opts.flags()
			if err != nil {
				return err
			}
			flags.NumberOfPackages = numberOfPackages
			return runner.Process(flags)
		},
	}

	cmd.Flags().IntVar(&numberOfPackages, "number_of_packages", 1,
		"Number of zip packages to create. Must not exceed the number of input files.")
	return cmd
}
