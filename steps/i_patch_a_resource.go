package steps

import (
	"context"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var APatch = stepdef.StepDefinition{
	Name: "a-patch",
	Text: "a patch called {reference}",
	Help: `A patch which will be applied in a subsequent step`,
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
	And a patch called mypatch
	  """application/merge-patch+json
          {
	    "data": {
	      "foo":"nobar"
	    }
	  }
	  """
	When I patch cm with mypatch
	Then cm jsonpath '{.data.foo}' should equal nobar`,
	// TODO StepArg:
	Function: func(ctx context.Context, ref string, ds *messages.DocString) (err error) {
		patchType := types.PatchType(ds.MediaType)
		rawPatch := []byte(ds.Content)

		patch := client.RawPatch(patchType, rawPatch)

		store.Save(ctx, ref, patch)

		return nil
	},
}

var IPatchAResource = stepdef.StepDefinition{
	Name: "i-patch",
	Text: "I patch {reference} with {reference}",
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
	Function: func(ctx context.Context, c client.WithWatch, ref *unstructured.Unstructured, patch client.Patch, opts []client.PatchOption) (err error) {
		return withRetry(ctx, func() error {
			return c.Patch(ctx, ref, patch, opts...)
		})
	},
}
