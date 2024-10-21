package cmd

import (
	"errors"
	"io"
	"os"

	"github.com/drrev/telehandler/pkg/work"
	"github.com/spf13/cobra"
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local <cmd> [args...]",
	Short: "Run the given command with cgroup and namespace enforcement.",
	Long: `Run the given command with cgroup and namespace enforcement.

The local command does not require a running Telehandler server, all commands
are executed in the local environment.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := work.NewExecutor(cgroupRoot)

		job := *work.NewJob("admin", args[0], args[1:])
		job, err := mgr.Start(job)
		if err != nil {
			return err
		}

		r, err := mgr.OpenReader(job.ID)
		if err != nil {
			return err
		}

		_, err = io.Copy(os.Stdout, r)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		return mgr.Wait(job.ID)
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
	localCmd.Flags().SetInterspersed(true)
}
