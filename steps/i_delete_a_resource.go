package steps

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	scheme.Default.MustAddToScheme(IDeleteAResource)
}

var IDeleteAResource = scheme.StepDefinition{
	Name: "i-delete",
	Text: "I delete <reference>",
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
	Parameters: []parameters.Parameter{parameters.Reference, parameters.DeleteOptions},
	Function: func(ctx context.Context, ref string, opts []client.DeleteOption) error {
		o := contextutils.LoadObject(ctx, ref)
		Expect(o).ShouldNot(BeNil(), ErrNoResource, ref)

		args := append([]interface{}{o}, opts)
		c := contextutils.MustGetClientFrom(ctx)
		Eventually(c.Delete).WithContext(ctx).WithArguments(args...).Should(Succeed(), "Failed to delete resource")

		return nil
	},
}
