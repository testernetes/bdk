package steps

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/models"
	"github.com/testernetes/bdk/register"
)

func init() {
	err := models.Scheme.Register(AResource)
	if err != nil {
		panic(err)
	}
}

var AResource = models.StepDefinition{
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
	Parameters: []models.Parameter{models.Reference, models.Manifest},
	Function: func(ctx context.Context, ref string, manifest *models.DocString) (err error) {
		u, err := manifest.GetUnstructured()
		Expect(err).ShouldNot(HaveOccurred())

		register.Save(ctx, ref, u)

		return nil
	},
}
