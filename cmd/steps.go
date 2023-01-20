/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
)

var stepHelpTemplate = `{{.Long}}

Step Definitions:
  {{.Short}}

Parameters:
  {{ parameters .Name }}

Examples:
{{.Example}}
`

func NewStepsCommand() *cobra.Command {
	cobra.AddTemplateFunc("parameters", printParameters)

	stepsCmd := &cobra.Command{
		Use:   "steps",
		Short: "View steps",
		Long:  "",
	}
	for _, s := range scheme.Default.GetStepDefs() {
		step := &cobra.Command{
			Use:     s.Name,
			Short:   s.Text,
			Long:    s.Help,
			Example: Examples(s.Examples),
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		step.SetHelpTemplate(stepHelpTemplate)
		stepsCmd.AddCommand(step)
	}
	return stepsCmd
}

func printParameters(stepName string) string {
	buf := bytes.NewBufferString("")
	for _, s := range scheme.Default.GetStepDefs() {
		if s.Name == stepName {
			for _, p := range s.Parameters {
				param, ok := p.(parameters.StringParameter)
				text := param.GetText()
				if !ok {
					text = "Additional Step Arguments"
				}
				fmt.Fprintf(buf, Examples("\n%s:"), text)
				fmt.Fprintf(buf, Parameter(p.GetShortHelp()))
				fmt.Fprintf(buf, Parameter(p.GetLongHelp()))
			}
		}
	}
	return buf.String()
}
