package cmd

import (
	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/spf13/cobra"
)

// stopCmd attempts to stop the given Job on the Telehandler server.
var stopCmd = &cobra.Command{
	Use:   "stop <job_id>",
	Short: "Attempts to stop the given job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := foremanClient.StopJob(cmd.Context(), &foremanpb.StopJobRequest{Name: args[0]})
		return err
	},
}

func init() {
	clientCmd.AddCommand(stopCmd)
}
