package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ICreateAResource = stepdef.StepDefinition{
	Name: "i-create",
	Text: "I create <reference>",
	Help: `Creates the referenced resource. Step will fail if the reference was not defined in a previous step.`,
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
	  """
	And I create cm
	  | field manager | example |
	Then within 1s cm jsonpath '{.metadata.uid}' should not be empty`,
	StepArg: stepdef.CreateOptions,
	Function: func(ctx context.Context, c client.WithWatch, reference *unstructured.Unstructured, opts []client.CreateOption) (err error) {
		return withRetry(ctx, func() error {
			return c.Create(ctx, reference, opts...)
		})
	},
}
