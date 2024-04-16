package simple

import (
	"fmt"
	"os"
	"strings"

	"atomicgo.dev/cursor"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/fatih/color"
	"github.com/testernetes/bdk/model"
	"github.com/testernetes/bdk/printers/utils"
	"github.com/testernetes/bdk/stepdef"
)

var colorFor = map[stepdef.Result]color.Attribute{
	stepdef.Passed:   color.FgGreen,
	stepdef.Skipped:  color.FgBlue,
	stepdef.Timedout: color.FgYellow,
	stepdef.Failed:   color.FgYellow,
	stepdef.Unknown:  color.FgRed,
}

type Printer struct {
}

func (p Printer) Print(events model.Events) {
	// always show color in github actions
	color.NoColor = color.NoColor && os.Getenv("GITHUB_ACTION") == ""

	for {
		event, more := <-events
		if !more {
			return
		}
		defer color.Unset()

		switch event.Type {
		case model.StartFeature:
			color.Set(color.FgWhite)
			fmt.Printf("%s: %s\n", event.Feature.Keyword, event.Feature.Name)
		case model.StartScenario:
			color.Set(color.FgWhite)
			utils.NewNormalizer("\n%s: %s", event.Scenario.Keyword, event.Scenario.Name).Indent(1).Println()
		case model.StartStep:
			s := fmt.Sprintf("%s%s", event.Step.Keyword, event.Step.Text)
			utils.NewNormalizer(s).Indent(2).Print()
			cursor.StartOfLine()
		case model.InProgressStep:
			color.Set(colorFor[event.StepResult.Result])
			s := fmt.Sprintf("%s%s", event.Step.Keyword, event.Step.Text)
			percent := int(event.StepResult.Progress*float64(len(s))) - 1
			if percent < 0 {
				percent = 1
			}
			utils.NewNormalizer(s[:percent]).Indent(2).Print()
			cursor.StartOfLine()
		case model.FinishStep:
			p.step(event.Step, event.Scenario.StepResults[event.Step])
		}
	}
}

func (p Printer) step(step *messages.Step, result stepdef.StepResult) {
	color.Set(colorFor[result.Result])
	utils.NewNormalizer("%s%s", step.Keyword, step.Text).Indent(2).Println()
	color.Unset()

	if step.DocString != nil {
		utils.NewNormalizer(step.DocString.Delimiter).Indent(2).Println()
		utils.NewNormalizer(step.DocString.Content).IndentTabs(2).Println()
		utils.NewNormalizer(step.DocString.Delimiter).Indent(2).Println()
	}
	if result.Result != stepdef.Passed {
		utils.NewNormalizer(strings.Join(result.Messages, "\n")).Indent(3).Println()
	}
	if result.Err != nil {
		utils.NewNormalizer(result.Err.Error()).Indent(3).Println()
	}
}
