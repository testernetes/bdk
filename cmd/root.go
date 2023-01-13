/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var formatter string
var plugins []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bdk",
	Short: "Behaviour Driven Kubernetes",
	Long:  "",
}

func NewRootCommand() *cobra.Command {
	rootCmd.PersistentFlags().StringSliceVarP(&plugins, "plugins", "p", []string{}, "Additional plugin step definitions")
	testCmd.Flags().StringVarP(&formatter, "format", "f", "simple", "the format printer")

	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(NewStepsCommand())

	return rootCmd
}
