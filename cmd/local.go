/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io"
	"os"

	"github.com/drrev/telehandler/pkg/work"
	"github.com/spf13/cobra"
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use: "local",

	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := work.NewExecutor(cgroupRoot)

		job := *work.NewJob("local", args[0], args[1:])
		if err := mgr.Start(job); err != nil {
			return err
		}

		r, err := mgr.Watch(job.ID)
		if err != nil {
			return err
		}

		io.Copy(os.Stdout, r)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
	localCmd.Flags().SetInterspersed(true)
}
