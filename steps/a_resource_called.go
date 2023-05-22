package steps

import (
	"context"

	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	scheme.Default.MustAddToScheme(AResource)
	scheme.Default.MustAddToScheme(AResourceFromFile)
}

var SaveObjectFunc = func(ctx context.Context, ref string, manifest *unstructured.Unstructured) (err error) {
	contextutils.SaveObject(ctx, ref, manifest)

	return nil
}

var AResourceFromFile = scheme.StepDefinition{
	Name: "a-resource-from-file",
	Text: `a <reference> from <filename>`,
	Help: `Assigns a reference to the resource given in the filename. This reference can be referred to
in future steps in the same scenario. JSON and YAML formats are accepted.`,
	Examples: `
	Given cm from config.yaml:`,
	Parameters: []parameters.Parameter{parameters.Reference, parameters.Filename},
	Function:   SaveObjectFunc,
}

var AResource = scheme.StepDefinition{
	Name: "a-resource",
	Text: `a resource called <reference>`,
	Help: `Assigns a reference to the resource given in the DocString. This reference can be referred to
in future steps in the same scenario. JSON and YAML formats are accepted.`,
	Examples: `
	Given a resource called cm:
	  """
	  apiVersion: v1
	  kind: ConfigMap
	  metadata:
	    name: example
	    namespace: default
	  data:
	    foo: bar
	  """`,
	Parameters: []parameters.Parameter{parameters.Reference, parameters.Manifest},
	Function:   SaveObjectFunc,
}
