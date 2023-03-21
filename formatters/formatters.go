package formatters

import (
	"fmt"

	"github.com/testernetes/bdk/formatters/configmap"
	"github.com/testernetes/bdk/formatters/debug"
	"github.com/testernetes/bdk/formatters/json"
	"github.com/testernetes/bdk/formatters/simple"
	"github.com/testernetes/bdk/model"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Printer interface {
	StartFeature(*model.Feature)
	FinishFeature(*model.Feature)
	StartScenario(*model.Feature, *model.Scenario)
	FinishScenario(*model.Feature, *model.Scenario)
	Print(*model.Feature)
}

func NewFormatter(name string) (Printer, error) {
	switch name {
	case "configmap":
		c, err := client.New(config.GetConfigOrDie(), client.Options{
			Scheme: scheme.Scheme,
		})
		if err != nil {
			return &configmap.Printer{}, err
		}
		return &configmap.Printer{Client: c}, nil
	case "json":
		return &json.Printer{}, nil
	case "simple":
		return &simple.Printer{}, nil
	case "debug":
		return &debug.Printer{}, nil
	}
	return nil, fmt.Errorf("not a valid printer")
}
