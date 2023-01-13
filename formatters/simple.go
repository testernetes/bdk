package formatters

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/testernetes/bdk/models"
)

func Simple(feature *models.Feature) {
	color.Set(color.FgWhite)
	printf(feature, fmt.Sprintf("%s: %s", feature.Keyword, feature.Name))
	for _, scenario := range feature.Scenarios {
		printf(scenario, fmt.Sprintf("%s: %s", scenario.Keyword, scenario.Name))
		for _, step := range scenario.Steps {
			switch step.Execution.Result {
			case models.Passed:
				color.Set(color.FgGreen)
			case models.Skipped:
				color.Set(color.FgBlue)
			case models.Timedout, models.Failed:
				color.Set(color.FgYellow)
			case models.Unknown, models.Interrupted:
				color.Set(color.FgRed)
			}

			fmt.Printf("    %s%s\n", step.Keyword, step.Text)
			if step.DocString != nil {
				fmt.Printf("    %s\n", step.DocString.Delimiter)
				fmt.Printf(indent(4, step.DocString.Content))
				fmt.Printf("    %s\n", step.DocString.Delimiter)
			}
			if step.Execution.Message != "Step Ran Successfully" {
				fmt.Printf(indent(4, step.Execution.Message))
			}
			if step.Execution.Err != nil {
				fmt.Printf(indent(4, step.Execution.Err))
			}
			color.Unset()
		}
	}
}

func printf(t interface{}, s interface{}) {
	switch t.(type) {
	case *models.Scenario:
		fmt.Printf(indent(2, s))
	default:
		fmt.Printf(indent(0, s))
	}
}

func indent(i int, s interface{}) string {
	ind := strings.Repeat(" ", i)
	return strings.TrimRight(strings.ReplaceAll(fmt.Sprintf("%s%v\n", ind, s), "\n", "\n"+ind), " ")
}
