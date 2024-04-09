/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/testernetes/bdk/formatters/utils"
	"github.com/testernetes/bdk/model"
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
	//cobra.AddTemplateFunc("parameters", printParameters)

	stepsCmd := &cobra.Command{
		Use:   "steps",
		Short: "View steps",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				//if args[0] == "print" {
				//	var w strings.Builder
				//	err := model.StepFunctions
				//	if err != nil {
				//		return err
				//	}
				//	fmt.Printf(w.String())
				//	return nil
				//}
			}
			cmd.Help()
			return nil
		},
	}
	for _, s := range *model.StepFunctions {
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
