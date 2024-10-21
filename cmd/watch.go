package cmd

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
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
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return watchJobOutput(cmd.Context(), args[0])
	},
}

func watchJobOutput(ctx context.Context, name string) error {
	s, err := foremanClient.WatchJobOutput(ctx, &foremanpb.WatchJobOutputRequest{Name: name})
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
