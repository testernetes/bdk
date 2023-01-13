package formatters

import (
	"github.com/kr/pretty"
	"github.com/testernetes/bdk/model"
)

func Debug(feature *model.Feature) {
	pretty.Println(feature)
}
