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
	Function: func(ctx context.Context, ref string, options []client.CreateOption) error {
		o := contextutils.LoadObject(ctx, ref)
		Expect(o).ShouldNot(BeNil(), ErrNoResource, ref)

		args := []interface{}{o}
		for _, opt := range options {
			args = append(args, opt)
		}

		c := contextutils.MustGetClientFrom(ctx)
		Eventually(c.Create).WithContext(ctx).WithArguments(args...).Should(Succeed(), "Failed to create resource")

		return nil
	},
}
