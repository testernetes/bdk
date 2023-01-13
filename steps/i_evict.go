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
	scheme.Default.MustAddToScheme(IEvict)
}

var IEvict = scheme.StepDefinition{
	Name: "i-evict",
	Text: "I evict <reference>",
	Help: "Evicts the referenced pod resource. Step will fail if the pod reference was not defined in a previous step.",
	Examples: `
	When I evict pod
	  | grace period seconds | 120 |`,
	Parameters: []parameters.Parameter{parameters.Reference, parameters.DeleteOptions},
	Function: func(ctx context.Context, ref string, options *arguments.DataTable) (err error) {
		pod := register.LoadPod(ctx, ref)
		Expect(pod).ShouldNot(BeNil(), ErrNoResource, ref)

		opts := client.DeleteOptionsFrom(pod, options)
		c := client.MustGetClientFrom(ctx)
		Eventually(c.Evict).WithContext(ctx).WithArguments(opts...).Should(Succeed(), "Failed to evict")

		return nil
	},
}
