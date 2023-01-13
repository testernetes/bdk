package steps

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/arguments"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/register"
	"github.com/testernetes/bdk/scheme"
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
	Function: func(ctx context.Context, ref string, options *arguments.DataTable) error {
		u := register.Load(ctx, ref)
		Expect(u).ShouldNot(BeNil(), ErrNoResource, ref)

		opts := client.DeleteOptionsFrom(u, options)
		c := client.MustGetClientFrom(ctx)
		Eventually(c.Delete).WithContext(ctx).WithArguments(opts...).Should(Succeed(), "Failed to delete resource")

		return nil
	},
}
