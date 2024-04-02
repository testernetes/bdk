package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IEvict = stepdef.StepDefinition{
	Name: "i-evict",
	Text: "I evict <reference>",
	Help: "Evicts the referenced pod resource. Step will fail if the pod reference was not defined in a previous step.",
	Examples: `
	When I evict pod
	  | grace period seconds | 120 |`,
	Function: func(ctx context.Context, c gkube.Client, ref *corev1.Pod, opts []client.DeleteOption) (err error) {
		return clientDo(ctx, func() error {
			return c.Evict(ctx, ref, opts...)
		})
	},
}
