package cmd

import (
	"github.com/drrev/telehandler/pkg/work"
	"github.com/spf13/cobra"
)

// reexecCmd is used to wrap the execution of a child process
// to configure cgroup and namespaces.
// This command should only be called by [work.Executor] as
// part of the [work.Job] execution process.
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
