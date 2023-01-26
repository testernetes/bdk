package formatters

import (
	"fmt"

	"github.com/testernetes/bdk/formatters/configmap"
	"github.com/testernetes/bdk/formatters/debug"
	"github.com/testernetes/bdk/formatters/json"
	"github.com/testernetes/bdk/formatters/simple"
	"github.com/testernetes/bdk/model"
)

type FeaturePrinter interface {
	Print(*model.Feature)
}

func NewFormatter(name string) (FeaturePrinter, error) {
	switch name {
	case "configmap":
		return &configmap.Printer{}, nil
	case "json":
		return &json.Printer{}, nil
	case "simple":
		return &simple.Printer{}, nil
	case "debug":
		return &debug.Printer{}, nil
	}
	return nil, fmt.Errorf("not a valid printer")
}
