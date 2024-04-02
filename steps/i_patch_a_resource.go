package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IPatchAResource = stepdef.StepDefinition{
	Name: "i-patch",
	Text: "I patch <reference>",
	Help: `Patches the referenced resource. Step will fail if the reference was not defined in a previous step.`,
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
	When I patch cm
	  | patch | {"data":{"foo":"nobar"}} |
	Then for at least 10s cm jsonpath '{.data.foo}' should equal nobar`,
	// TODO StepArg:
	Function: func(ctx context.Context, c client.Client, ref *unstructured.Unstructured, opts []client.PatchOption) (err error) {
		var p client.Patch // TODO
		return clientDo(ctx, func() error {
			return c.Patch(ctx, ref, p, opts...)
		})
	},
}
