package simple

import (
	"fmt"
	"os"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/fatih/color"
	"github.com/testernetes/bdk/model"
	"github.com/testernetes/bdk/printers/utils"
)

type Printer struct {
	indent int
}

func (p Printer) Print(events model.Events) {
	for {
		event, more := <-events
		fmt.Printf("%s\n", event.Type)
		if !more {
			return
		}

		switch event.Type {
		case model.FinishScenario:
			p.PrintScenario(event.Scenario)
		}
	}
}

func (p Printer) feature(feature *model.Feature) {
	color.Set(color.FgWhite)
	defer color.Unset()
	fmt.Printf("%s: %s\n", feature.Keyword, feature.Name)
}

func (p Printer) scenario(scenario *model.Scenario) {
	color.Set(color.FgWhite)
	defer color.Unset()
	//utils.NewNormalizer("\n%s: %s", scenario.Keyword, scenario.Name).Indent(1).Print()
}

var colorFor = map[model.Result]color.Attribute{
	model.Passed:   color.FgGreen,
	model.Skipped:  color.FgBlue,
	model.Timedout: color.FgYellow,
	model.Failed:   color.FgYellow,
	model.Unknown:  color.FgRed,
}

func (p Printer) step(step *messages.Step, result model.StepResult) {
	color.Set(colorFor[result.Result])
	utils.NewNormalizer("%s%s", step.Keyword, step.Text).Indent(2).Print()
	color.Unset()

	if step.DocString != nil {
		utils.NewNormalizer(step.DocString.Delimiter).Indent(2).Print()
		utils.NewNormalizer(step.DocString.Content).IndentTabs(2).Print()
		utils.NewNormalizer(step.DocString.Delimiter).Indent(2).Print()
	}
	if result.Message != "Step Ran Successfully" {
		utils.NewNormalizer(result.Message).Indent(3).Print()
	}
	if result.Err != nil {
		utils.NewNormalizer(result.Err.Error()).Indent(3).Print()
	}
}

func (p Printer) StartFeature(feature *model.Feature) {
	// always show color in github actions
	color.NoColor = color.NoColor && os.Getenv("GITHUB_ACTION") == ""
	p.feature(feature)
}

func (p Printer) PrintScenario(scenario *model.Scenario) {
	p.scenario(scenario)
	for step, result := range scenario.StepResults {
		p.step(step, result)
	}
}

func (p Printer) StartScenario(feature *model.Feature, scenario *model.Scenario) {}
func (p Printer) FinishFeature(feature *model.Feature)                           {}
