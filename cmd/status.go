package cmd

import (
	"log/slog"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/spf13/cobra"
)

// statusCmd attempts to get the current status of a given Job on the Telehandler server.
var statusCmd = &cobra.Command{
	Use:   "status <job_id>",
	Short: "Attempts to status the given job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := foremanClient.GetJobStatus(cmd.Context(), &foremanpb.GetJobStatusRequest{Name: args[0]})

		attrs := []slog.Attr{
			slog.String("name", status.GetName()),
			slog.Any("state", status.GetState()),
			slog.Time("start", status.GetStartTime().AsTime()),
		}
		if status.GetEndTime().IsValid() {
			attrs = append(attrs, slog.Time("end", status.GetEndTime().AsTime()))
			attrs = append(attrs[:2], attrs[1:]...)
			attrs[2] = slog.Duration("duration", status.GetEndTime().AsTime().Sub(status.GetStartTime().AsTime()))
		}
		if status.GetState() != foremanpb.JobState_JOB_STATE_RUNNING {
			attrs = append(attrs, slog.Int("exit_code", int(status.GetExitCode())))
		}

		slog.Info("Status", slog.Any("job", attrs))
		return err
	},
}

func init() {
	clientCmd.AddCommand(statusCmd)
}
