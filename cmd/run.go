package cmd

import (
	"log/slog"
	"os"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
)

// runCmd executes the given command using a Telehandler server.
// To simplify usage, this command automatically streams Job output.
var runCmd = &cobra.Command{
	Use:   "run <command> [args...]",
	Short: "Run a Linux command using a Telehandler server",
	Long: `Run a Linux command using a Telehandler server.
	
Commands that include args should signal the end of args for this command with --
For example: run -- bash -c echo hello"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := foremanClient.StartJob(cmd.Context(), &foremanpb.StartJobRequest{Command: args[0], Args: args[1:]})
		if err != nil {
			st := status.Convert(err)
			slog.Error("Failed to start job", slog.Any("code", st.Code()), slog.String("error", st.Message()))
		}
		if err := os.WriteFile(jidFile, []byte(resp.GetId()), 0o644); err != nil {
			slog.Error("Failed to write jidfile", slog.Any("error", err))
		}
		return watchJobOutput(cmd.Context(), resp.GetId())
	},
}

func init() {
	clientCmd.AddCommand(runCmd)
	runCmd.Flags().SetInterspersed(true)
}
