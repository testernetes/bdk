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
	err := models.Scheme.Register(IEvict)
	if err != nil {
		panic(err)
	}
}

var IEvict = models.StepDefinition{
	Name: "i-evict",
	Text: "I evict <reference>",
	Help: "Evicts the referenced pod resource. Step will fail if the pod reference was not defined in a previous step.",
	Examples: `
	When I evict pod
	  | grace period seconds | 120 |`,
	Parameters: []models.Parameter{models.Reference, models.DeleteOptions},
	Function: func(ctx context.Context, ref string, options *messages.DataTable) (err error) {
		pod := register.LoadPod(ctx, ref)
		Expect(pod).ShouldNot(BeNil(), ErrNoResource, ref)

		opts := client.DeleteOptionsFrom(pod, options)
		c := client.MustGetClientFrom(ctx)
		Eventually(c.Evict).WithContext(ctx).WithArguments(opts...).Should(Succeed(), "Failed to evict")

		return nil
	},
}
