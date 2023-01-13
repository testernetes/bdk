package formatters

import (
	"fmt"

	"github.com/testernetes/bdk/models"
	"sigs.k8s.io/yaml"
)

func JSON(feature *models.Feature) {
	out, _ := yaml.Marshal(feature)
	fmt.Printf("%s\n", out)
}
