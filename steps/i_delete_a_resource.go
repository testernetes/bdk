package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IDelete = stepdef.StepDefinition{
	Name: "i-delete",
	Text: "^I delete {reference}$",
	Help: "Deletes the referenced resource. Step will fail if the reference was not defined in a previous step.",
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
	And I delete cm
	  | grace period seconds | 30         |
	  | propagation policy   | Foreground |`,
	StepArg: stepdef.DeleteOptions,
	Function: func(ctx context.Context, t *stepdef.T, ref *unstructured.Unstructured, opts []client.DeleteOption) error {
		return t.WithRetry(ctx, func() error {
			return t.Client.Delete(ctx, ref, opts...)
		}, stepdef.RetryK8sError)
	},
}
