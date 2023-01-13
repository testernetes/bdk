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
	err := models.Scheme.Register(ICreateAResource)
	if err != nil {
		panic(err)
	}
}

var ICreateAResource = models.StepDefinition{
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
	Parameters: []models.Parameter{models.Reference, models.CreateOptions},
	Function: func(ctx context.Context, ref string, options *messages.DataTable) error {
		u := register.Load(ctx, ref)
		Expect(u).ShouldNot(BeNil(), ErrNoResource, ref)

		opts := client.CreateOptionsFrom(u, options)
		c := client.MustGetClientFrom(ctx)
		Eventually(c.Create).WithContext(ctx).WithArguments(opts...).Should(Succeed(), "Failed to create resource")

		return nil
	},
}
