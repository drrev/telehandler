package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// watchCmd streams output from the given Job.
// Multiple streams can be active for a Job at any given time.
var watchCmd = &cobra.Command{
	Use:   "watch [job_id]",
	Short: "Watch the output of a job",
	Long: `Watch the output of a job starting at the beginning of execution through process termination.
	
Jobs do not need to be running to watch output.
If the job is not running, all historical output from process start to finish is retrieved.`,
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
		return watchJobOutput(cmd.Context(), args[0])
	},
}

func watchJobOutput(ctx context.Context, id string) error {
	s, err := foremanClient.WatchJobOutput(ctx, &foremanpb.WatchJobOutputRequest{Id: id})
	if err != nil {
		log.Fatal(err)
	}

	for {
		out, err := s.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if data := out.GetData(); len(data) > 0 {
			os.Stdout.Write(data)
		}
	}
}

func init() {
	clientCmd.AddCommand(watchCmd)
}
