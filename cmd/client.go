package cmd

import (
	"fmt"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/internal/auth"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	grpcConn       *grpc.ClientConn
	foremanClient  foremanpb.ForemanServiceClient
	jidFile        = "job_id"
	clientCertPath = "ssl/client.pem"
	clientKeyPath  = "ssl/client-key.pem"
	teleServer     = "localhost:6443"
	userName       = ""
)

// clientCmd is a meta command to group all client commands together.
//
// Execution of a client command automatically creates and
// connects a [foremanpb.ForemanServiceClient] automatically.
// Client subcommands should use the created foremanClient for RPCs.
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "client is used to run subcommands over gRPC",

	// Eagerly setup client, assume the command will work
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		tlsConfig, cn, err := auth.LoadClientTLS(clientCertPath, clientKeyPath, caCertPath)
		if err != nil {
			return fmt.Errorf("failed to load TLS config: %w", err)
		}
		userName = cn

		grpcConn, err = grpc.NewClient(teleServer, grpc.WithTransportCredentials(tlsConfig))
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
		foremanClient = foremanpb.NewForemanServiceClient(grpcConn)

		return nil
	},
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		if grpcConn != nil {
			return grpcConn.Close()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.PersistentFlags().StringVarP(&teleServer, "server", "s", "localhost:6443", "Address of a Telehandler server")
	clientCmd.PersistentFlags().StringVarP(&clientCertPath, "cert", "c", clientCertPath, "Client cert path")
	clientCmd.PersistentFlags().StringVarP(&clientKeyPath, "key", "k", clientKeyPath, "Client key path")
	clientCmd.PersistentFlags().StringVarP(&jidFile, "jidfile", "j", jidFile, "A file to write the ID of the Job.")
}
