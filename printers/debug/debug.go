package debug

import (
	"github.com/kr/pretty"
	"github.com/testernetes/bdk/model"
)

type Printer struct{}

func (p Printer) Print(feature *model.Feature) {
	pretty.Println(feature)
}

func (p Printer) StartFeature(feature *model.Feature)                             {}
func (p Printer) FinishFeature(feature *model.Feature)                            {}
func (p Printer) StartScenario(feature *model.Feature, scenario *model.Scenario)  {}
func (p Printer) FinishScenario(feature *model.Feature, scenario *model.Scenario) {}
