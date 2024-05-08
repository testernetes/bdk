package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IGet = stepdef.StepDefinition{
	Name: "i-get",
	Text: "^I get {reference}$",
	Help: `Gets the referenced resource. Step will fail if the reference was not defined in a previous step.`,
	Examples: `
	Given a cm from file blah.yaml
	And I get cm
	Then cm jsonpath '{.metadata.uid}' should not be empty`,
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, t *stepdef.T, ref *unstructured.Unstructured) error {
		return t.WithRetry(ctx, func() error {
			return t.Client.Get(ctx, client.ObjectKeyFromObject(ref), ref)
		}, stepdef.RetryK8sError)
	},
}
