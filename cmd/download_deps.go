package cmd

import "github.com/spf13/cobra"

func newDownloadDepsCmd(rootOpts *rootOptions, runner Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "download_deps",
		Short: "Download the dependencies (maps) of the input replays without processing them.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags, err := rootOpts.baseFlags()
			if err != nil {
				return err
			}
			flags.OnlyDependencyDownload = true
			return runner.DownloadDeps(flags)
		},
	}
}
