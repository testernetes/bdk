/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"bytes"

	"github.com/spf13/cobra"
	"github.com/testernetes/bdk/stepdef"
)

var matcherHelpTemplate = `{{.Long}}

Matcher:
  {{.Short}}

Parameters:
  {{ matcherParameters .Name }}
`

func NewMatchersCommand() *cobra.Command {
	cobra.AddTemplateFunc("matcherParameters", printMatcherParameters)

	matchersCmd := &cobra.Command{
		Use:   "matchers",
		Short: "View matchers",
		Long:  "",
	}
	for _, m := range stepdef.Matchers {
		if m.Name == "" {
			continue
		}
		matcher := &cobra.Command{
			Use:   m.Name,
			Short: m.Text,
			Long:  m.Help,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		matcher.SetHelpTemplate(matcherHelpTemplate)
		matchersCmd.AddCommand(matcher)
	}
	return matchersCmd
}

func printMatcherParameters(name string) string {
	buf := bytes.NewBufferString("")
	for _, m := range stepdef.Matchers {
		if m.Name == name {
			m.PrintHelp()
		}
	}
	return buf.String()
}
