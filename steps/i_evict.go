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
	Function: func(ctx context.Context, ref string, opts []client.DeleteOption) (err error) {
		pod := contextutils.LoadPod(ctx, ref)
		Expect(pod).ShouldNot(BeNil(), ErrNoResource, ref)

		args := []interface{}{pod}
		for _, opt := range opts {
			args = append(args, opt)
		}

		c := contextutils.MustGetClientFrom(ctx)
		Eventually(c.Evict).WithContext(ctx).WithArguments(args...).Should(Succeed(), "Failed to evict")

		return nil
	},
}
