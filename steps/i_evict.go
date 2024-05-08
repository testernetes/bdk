package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IEvict = stepdef.StepDefinition{
	Name: "i-evict",
	Text: "^I evict {reference}$",
	Help: "Evicts the referenced pod resource. Step will fail if the pod reference was not defined in a previous step.",
	Examples: `
	When I evict pod
	  | grace period seconds | 120 |`,
	StepArg: stepdef.DeleteOptions,
	Function: func(ctx context.Context, t *stepdef.T, pod *corev1.Pod, opts []client.DeleteOption) (err error) {
		deleteOptions := &client.DeleteOptions{}
		for _, opt := range opts {
			opt.ApplyToDelete(deleteOptions)
		}

		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name: pod.Name,
			},
			DeleteOptions: deleteOptions.AsDeleteOptions(),
		}

		return t.WithRetry(ctx, func() error {
			return t.Client.SubResource("eviction").Create(ctx, pod, eviction)
		}, stepdef.RetryK8sError)
	},
}
