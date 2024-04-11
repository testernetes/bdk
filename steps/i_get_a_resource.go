package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IGet = stepdef.StepDefinition{
	Name: "i-get",
	Text: "I get {reference}",
	Help: `Gets the referenced resource. Step will fail if the reference was not defined in a previous step.`,
	Examples: `
	Given a cm from file blah.yaml
	And I get cm
	Then cm jsonpath '{.metadata.uid}' should not be empty`,
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, c client.WithWatch, ref *unstructured.Unstructured) error {
		return withRetry(ctx, func() error {
			return c.Get(ctx, client.ObjectKeyFromObject(ref), ref)
		})
	},
}
