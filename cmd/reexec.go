package cmd

import (
	"github.com/drrev/telehandler/pkg/work"
	"github.com/spf13/cobra"
)

// reexecCmd represents the reexec command.
var reexecCmd = &cobra.Command{
	Hidden: true,
	Use:    "reexec",
	RunE: func(_ *cobra.Command, args []string) error {
		return work.Reexec(cgroupRoot, args)
	},
}

func init() {
	rootCmd.AddCommand(reexecCmd)
	reexecCmd.Flags().SetInterspersed(true)
}
