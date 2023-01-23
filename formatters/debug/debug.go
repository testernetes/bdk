package debug

import (
	"github.com/kr/pretty"
	"github.com/testernetes/bdk/model"
)

type Printer struct{}

func (p Printer) Print(feature *model.Feature) {
	pretty.Println(feature)
}
