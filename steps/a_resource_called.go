package steps

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/arguments"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/register"
	"github.com/testernetes/bdk/scheme"
)

func init() {
	scheme.Default.MustAddToScheme(AResource)
}

var AResource = scheme.StepDefinition{
	Name: "a-resource",
	Text: `a resource called <reference>`,
	Help: `Assigns a reference to the resource given the in the DocString. This reference can be referred to
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
	Function: func(ctx context.Context, ref string, manifest *arguments.DocString) (err error) {
		u, err := manifest.GetUnstructured()
		Expect(err).ShouldNot(HaveOccurred())

		register.Save(ctx, ref, u)

		return nil
	},
}
