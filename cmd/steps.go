/*
Copyright © 2023 Matt Simons
*/
package cmd

import (
	"github.com/spf13/cobra"
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
		cmdStep := &cobra.Command{
			Use:     s.Name,
			Short:   s.Text,
			Long:    s.Help,
			Example: s.Examples,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		}
		cmdStep.SetHelpFunc(func(c *cobra.Command, args []string) {
			s.PrintHelp()
		})
		stepsCmd.AddCommand(cmdStep)
	}
	return stepsCmd
}
