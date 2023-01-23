package json

import (
	"fmt"

	"github.com/testernetes/bdk/model"
	"sigs.k8s.io/yaml"
)

type Printer struct{}

func (p Printer) Print(feature *model.Feature) {
	out, _ := yaml.Marshal(feature)
	fmt.Printf("%s\n", out)
}
