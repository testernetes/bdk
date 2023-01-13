package formatters

import (
	"fmt"

	"github.com/testernetes/bdk/models"
)

func Print(name string, feature *models.Feature) {
	switch name {
	case "json":
		JSON(feature)
	case "simple":
		Simple(feature)
	case "debug":
		Debug(feature)
	default:
		fmt.Println("not a valid printer")
	}
}
