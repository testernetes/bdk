package formatters

import (
	"fmt"

	"github.com/testernetes/bdk/model"
	"sigs.k8s.io/yaml"
)

func JSON(feature *model.Feature) {
	out, _ := yaml.Marshal(feature)
	fmt.Printf("%s\n", out)
}
