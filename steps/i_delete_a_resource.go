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
	err := models.Scheme.Register(IDeleteAResource)
	if err != nil {
		panic(err)
	}
}

var IDeleteAResource = models.StepDefinition{
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
	Parameters: []models.Parameter{models.Reference, models.DeleteOptions},
	Function: func(ctx context.Context, ref string, options *messages.DataTable) error {
		u := register.Load(ctx, ref)
		Expect(u).ShouldNot(BeNil(), ErrNoResource, ref)

		opts := client.DeleteOptionsFrom(u, options)
		c := client.MustGetClientFrom(ctx)
		Eventually(c.Delete).WithContext(ctx).WithArguments(opts...).Should(Succeed(), "Failed to delete resource")

		return nil
	},
}
