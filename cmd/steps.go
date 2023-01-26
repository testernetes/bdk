/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/testernetes/bdk/formatters/utils"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				if args[0] == "print" {
					var w strings.Builder
					err := scheme.Default.GenMarkdown(&w)
					if err != nil {
						return err
					}
					fmt.Printf(w.String())
					return nil
				}
			}
			cmd.Help()
			return nil
		},
	}
	for _, s := range scheme.Default.GetStepDefs() {
		step := &cobra.Command{
			Use:     s.Name,
			Short:   s.Text,
			Long:    s.Help,
			Example: utils.Examples(s.Examples),
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
				fmt.Fprintf(buf, utils.Examples("\n%s:"), text)
				fmt.Fprintf(buf, utils.Parameter(p.GetShortHelp()))
				fmt.Fprintf(buf, utils.Parameter(p.GetLongHelp()))
			}
		}
	}
	return buf.String()
}
