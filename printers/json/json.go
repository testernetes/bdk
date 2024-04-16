package json

import (
	"fmt"

	"github.com/testernetes/bdk/model"
	"sigs.k8s.io/yaml"
)

type Printer struct{}

func (p Printer) Print(events model.Events) {
	var features []*model.Feature
	for {
		event, more := <-events
		if !more {
			out, err := yaml.Marshal(features)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("%s\n", out)
			return
		}

		switch event.Type {
		case model.FinishFeature:
			features = append(features, event.Feature)
		}
	}
}
