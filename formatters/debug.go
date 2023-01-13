package formatters

import (
	"github.com/kr/pretty"
	"github.com/testernetes/bdk/models"
)

func Debug(feature *models.Feature) {
	pretty.Println(feature)
}
