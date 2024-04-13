package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var AResourceFromFile = stepdef.StepDefinition{
	Name:     "a-resource-from-file",
	Text:     "^a {reference} from {filename}$",
	Function: SaveObjectFunc,
	StepArg:  stepdef.NoStepArg,
	Help: `Assigns a reference to the resource given in the filename. This reference can be referred to
in future steps in the same scenario. JSON and YAML formats are accepted.`,
	Examples: `
	Given cm from config.yaml:`,
}

var AResource = stepdef.StepDefinition{
	Name:     "a-resource",
	Text:     "^a resource called {reference}$",
	Function: SaveObjectFunc,
	StepArg:  stepdef.Manifest,
	Help: `Assigns a reference to the resource given in the DocString. This reference can be referred to
in future steps in the same scenario. JSON and YAML formats are accepted.`,
	Examples: `Given a resource called cm:
	  """
	  apiVersion: v1
	  kind: ConfigMap
	  metadata:
	    name: example
	    namespace: default
	  data:
	    foo: bar
	  """`,
}

var SaveObjectFunc = func(ctx context.Context, out *unstructured.Unstructured, in *unstructured.Unstructured) (err error) {
	*out = *in.DeepCopy()
	return nil
}
