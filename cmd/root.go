/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var plugins []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bdk",
	Short: "Behaviour Driven Kubernetes",
	Long:  "",
}

func NewRootCommand() *cobra.Command {
	rootCmd.PersistentFlags().StringSliceVarP(&plugins, "plugins", "p", []string{}, "Additional plugin step definitions")

	rootCmd.AddCommand(NewTestCommand())
	rootCmd.AddCommand(NewStepsCommand())
	rootCmd.AddCommand(NewMatchersCommand())

	err := doc.GenMarkdownTree(rootCmd, "/tmp")
	if err != nil {
		log.Fatal(err)
	}

	return rootCmd
}
