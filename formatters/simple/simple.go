package simple

import (
	"fmt"
	"os"

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
	utils.NewNormalizer("\n%s: %s", scenario.Keyword, scenario.Name).Indent(1).Print()
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
	utils.NewNormalizer("%s%s", step.Keyword, step.Text).Trim().Indent(2).Print()
	color.Unset()
	if step.DocString != nil {
		utils.NewNormalizer(step.DocString.Delimiter).Indent(2).Print()
		utils.NewNormalizer(step.DocString.Content).IndentTabs(2).Print()
		utils.NewNormalizer(step.DocString.Delimiter).Indent(2).Print()
	}
	if step.Execution.Message != "Step Ran Successfully" {
		utils.NewNormalizer(step.Execution.Message).Indent(3).Print()
	}
	if step.Execution.Err != nil {
		utils.NewNormalizer(step.Execution.Err.Error()).Indent(3).Print()
	}
}

func (p Printer) StartFeature(feature *model.Feature) {
	// always show color in github actions
	color.NoColor = color.NoColor && os.Getenv("GITHUB_ACTION") == ""
	p.feature(feature)
}

func (p Printer) FinishScenario(feature *model.Feature, scenario *model.Scenario) {
	p.scenario(scenario)
	for _, step := range scenario.Steps {
		p.step(step)
	}
}

func (p Printer) StartScenario(feature *model.Feature, scenario *model.Scenario) {}
func (p Printer) FinishFeature(feature *model.Feature)                           {}
func (p Printer) Print(feature *model.Feature)                                   {}
