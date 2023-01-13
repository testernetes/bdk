package steps

import (
	"context"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/models"
	"github.com/testernetes/bdk/register"
)

func init() {
	err := models.Scheme.Register(IPatchAResource)
	if err != nil {
		panic(err)
	}
}

var IPatchAResource = models.StepDefinition{
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
	Parameters: []models.Parameter{models.Reference, models.PatchOptions},
	Function: func(ctx context.Context, ref string, options *messages.DataTable) (err error) {
		u := register.Load(ctx, ref)
		Expect(u).ShouldNot(BeNil(), ErrNoResource, ref)

		args := client.PatchOptionsFrom(u, options)
		c := client.MustGetClientFrom(ctx)
		Eventually(c.Patch).WithContext(ctx).WithArguments(args...).Should(Succeed())

		return nil
	},
}
