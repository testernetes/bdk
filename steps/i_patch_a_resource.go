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
	scheme.Default.MustAddToScheme(IPatchAResource)
}

var IPatchAResource = scheme.StepDefinition{
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
	Parameters: []parameters.Parameter{parameters.Reference, parameters.PatchOptions},
	Function: func(ctx context.Context, ref string, opts []client.PatchOption) (err error) {
		o := contextutils.LoadObject(ctx, ref)
		Expect(o).ShouldNot(BeNil(), ErrNoResource, ref)

		args := append([]interface{}{o}, opts)
		c := contextutils.MustGetClientFrom(ctx)
		Eventually(c.Patch).WithContext(ctx).WithArguments(args...).Should(Succeed())

		return nil
	},
}
