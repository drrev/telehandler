package cmd

import (
	"fmt"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// stopCmd attempts to stop the given Job on the Telehandler server.
var stopCmd = &cobra.Command{
	Use:   "stop <job_id>",
	Short: "Attempts to stop the given job",
	Args: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("invalid args expected <job_id>")
		}

		if _, err := uuid.Parse(args[0]); err != nil {
			return fmt.Errorf("invalid job id '%v'", args[0])
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := foremanClient.StopJob(cmd.Context(), &foremanpb.StopJobRequest{Id: args[0]})
		return err
	},
}

func init() {
	clientCmd.AddCommand(stopCmd)
}
