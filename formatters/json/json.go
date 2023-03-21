package json

import (
	"fmt"

	"github.com/testernetes/bdk/model"
	"sigs.k8s.io/yaml"
)

type Printer struct{}

func (p Printer) Print(feature *model.Feature) {
	out, err := yaml.Marshal(feature)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", out)
}

func (p Printer) StartFeature(feature *model.Feature)                             {}
func (p Printer) FinishFeature(feature *model.Feature)                            {}
func (p Printer) StartScenario(feature *model.Feature, scenario *model.Scenario)  {}
func (p Printer) FinishScenario(feature *model.Feature, scenario *model.Scenario) {}
