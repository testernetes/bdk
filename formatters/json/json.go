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
