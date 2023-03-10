package simple

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/testernetes/bdk/formatters/utils"
	"github.com/testernetes/bdk/model"
)

type Printer struct{}

func (p Printer) feature(feature *model.Feature) {
	color.Set(color.FgWhite)
	defer color.Unset()
	fmt.Printf("%s: %s\n", feature.Keyword, feature.Name)
}

func (p Printer) scenario(scenario *model.Scenario) {
	color.Set(color.FgWhite)
	defer color.Unset()
	utils.NewNormalizer("%s: %s", scenario.Keyword, scenario.Name).Indent(1).Print()
}

func (p Printer) step(step *model.Step) {
	switch step.Execution.Result {
	case model.Passed:
		color.Set(color.FgGreen)
	case model.Skipped:
		color.Set(color.FgBlue)
	case model.Timedout, model.Failed:
		color.Set(color.FgYellow)
	case model.Unknown, model.Interrupted:
		color.Set(color.FgRed)
	}
	defer color.Unset()

	utils.NewNormalizer("%s%s", step.Keyword, step.Text).Indent(2).Print()
	if step.DocString != nil {
		utils.NewNormalizer(step.DocString.Delimiter).Indent(2).Print()
		utils.NewNormalizer(step.DocString.Content).IndentTabs(2).Print()
		utils.NewNormalizer(step.DocString.Delimiter).Indent(2).Print()
	}
	if step.Execution.Message != "Step Ran Successfully" {
		utils.NewNormalizer(step.Execution.Message).Indent(2).Print()
	}
	if step.Execution.Err != nil {
		utils.NewNormalizer(step.Execution.Err.Error()).Indent(2).Print()
	}
}

func (p Printer) Print(feature *model.Feature) {
	p.feature(feature)
	for _, scenario := range feature.Scenarios {
		p.scenario(scenario)
		for _, step := range scenario.Steps {
			p.step(step)
		}
	}
}
