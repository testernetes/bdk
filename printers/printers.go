package printers

import (
	"fmt"

	"github.com/testernetes/bdk/model"
	"github.com/testernetes/bdk/printers/simple"
)

var Printers = map[string]Printer{
	"simple": &simple.Printer{},
	//"configmap": &configmap.Printer{},
	//"json":      &json.Printer{},
	//"debug":     &debug.Printer{},
}

type Printer interface {
	Print(model.Events)
}

func NewPrinter(name string) (Printer, error) {
	if p, ok := Printers[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("no printer called %s", name)
}
