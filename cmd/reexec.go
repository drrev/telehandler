package cmd

import (
	"os/signal"
	"syscall"

	"github.com/drrev/telehandler/pkg/work"
	"github.com/spf13/cobra"
)

// reexecCmd is used to wrap the execution of a child process
// to configure cgroup and namespaces.
// This command should only be called by [work.Executor] as
// part of the [work.Job] execution process.
var reexecCmd = &cobra.Command{
	Hidden:        true,
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           "reexec",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// intercept signals for graceful shutdown
		basectx, cancel := signal.NotifyContext(
			ctx,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGQUIT,
			syscall.SIGTERM,
		)
		defer cancel()

		return work.Reexec(basectx, cgroupRoot, args)
	},
}

func init() {
	rootCmd.AddCommand(reexecCmd)
	reexecCmd.Flags().SetInterspersed(true)
}
