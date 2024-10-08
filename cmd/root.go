package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cert       = "ssl/server.pem"
	key        = "ssl/server-key.pem"
	root       = "ssl/root.pem"
	cgroupRoot = "/sys/fs/cgroup"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "telehandler",
	Short: "Telehandler is a simple service that is used to start, stop, query status, and watch the output of an arbitrary Linux process over gRPC.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cert, "cert", "c", cert, "Server cert path")
	rootCmd.PersistentFlags().StringVarP(&key, "key", "k", key, "Server key path")
	rootCmd.PersistentFlags().StringVarP(&root, "root", "r", root, "Root CA cert path")
	rootCmd.PersistentFlags().StringVar(&cgroupRoot, "cgroup-root", cgroupRoot, "Path to cgroup v2 mount")

}
