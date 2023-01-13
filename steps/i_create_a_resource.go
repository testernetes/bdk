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
	scheme.Default.MustAddToScheme(ICreateAResource)
}

var ICreateAResource = scheme.StepDefinition{
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
	Parameters: []parameters.Parameter{parameters.Reference, parameters.CreateOptions},
	Function: func(ctx context.Context, ref string, options *arguments.DataTable) error {
		u := register.Load(ctx, ref)
		Expect(u).ShouldNot(BeNil(), ErrNoResource, ref)

		opts := client.CreateOptionsFrom(u, options)
		c := client.MustGetClientFrom(ctx)
		Eventually(c.Create).WithContext(ctx).WithArguments(opts...).Should(Succeed(), "Failed to create resource")

		return nil
	},
}
