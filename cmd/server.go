package cmd

import (
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	_ "google.golang.org/grpc/encoding/gzip"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/internal/auth"
	"github.com/drrev/telehandler/internal/foreman"
	"github.com/drrev/telehandler/pkg/work"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	listenProtocol = "tcp"
	listenAddress  = ":6443"
	serverCertPath = "ssl/server.pem"
	serverKeyPath  = "ssl/server-key.pem"
)

// serverCmd runs a [foremanpb.ForemanService].
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a gRPC server for running and managing jobs",
	RunE: func(cmd *cobra.Command, _ []string) error {
		tlsConfig, err := auth.LoadServerTLS(serverCertPath, serverKeyPath, caCertPath)
		if err != nil {
			return fmt.Errorf("failed to load tls config: %w", err)
		}

		server := grpc.NewServer(
			grpc.Creds(tlsConfig),
			grpc.InTapHandle(auth.Tap),
			grpc.UnaryInterceptor(auth.UnaryServerInterceptor),
			grpc.StreamInterceptor(auth.ServerStreamInterceptor),
		)

		foremanpb.RegisterForemanServiceServer(server, foreman.NewService(work.NewExecutor(cgroupRoot)))

		// intercept signals for graceful shutdown
		basectx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		listener, err := net.Listen(listenProtocol, listenAddress)
		if err != nil {
			return fmt.Errorf("failed to open listen on %s: %w", listenAddress, err)
		}

		slog.LogAttrs(basectx, slog.LevelInfo, "Listening", slog.String("addr", listenAddress))

		errc := make(chan error, 1)
		defer close(errc)
		go func() {
			if err := server.Serve(listener); err != nil {
				errc <- err
			}
		}()

		select {
		case err := <-errc:
			return err
		case <-basectx.Done():
		}

		// give time to gracefully stop, then kill
		go func() {
			<-time.After(10 * time.Second)
			server.Stop()
		}()

		server.GracefulStop()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&listenProtocol, "protocol", "p", listenProtocol, "protocol for incoming connections")
	serverCmd.Flags().StringVarP(&listenAddress, "listen", "l", listenAddress, "ip:port to listen on for incoming connections")
	serverCmd.PersistentFlags().StringVarP(&serverCertPath, "cert", "c", serverCertPath, "Server cert path")
	serverCmd.PersistentFlags().StringVarP(&serverKeyPath, "key", "k", serverKeyPath, "Server key path")
}
