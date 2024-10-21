//go:build docs
// +build docs

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// genCmd generates a markdown tree in docs/cli for all
// non-hidden commands registered to rootCmd.
var genCmd = &cobra.Command{
	Hidden: true,
	Use:    "gen",
	Short:  "Generates Markdown documentation for the Telehandler CLI and places it into ./docs/cli",
	RunE: func(_ *cobra.Command, _ []string) error {
		return doc.GenMarkdownTree(rootCmd, "./docs/cli")
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
}
