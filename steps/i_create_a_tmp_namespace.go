package steps

import (
	"context"
	"fmt"

	"github.com/testernetes/bdk/stepdef"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ICreateATmpNamespace = stepdef.StepDefinition{
	Name: "i-create-a-tmp-namespace",
	Text: "^I create a temporary namespace called {var}$",
	Help: `Creates a namespace of the same name as reference with some random characters`,
	Examples: `
	Given I create a temporary namespace called scenario
	Given a resource called cm
	  """
	  apiVersion: v1
	  kind: ConfigMap
	  metadata:
	    name: example
	    namespace: ${scenario}
	  data:
	    foo: bar
	  """
	And I create cm
	Then within 1s cm jsonpath '{.metadata.namespace}' should match regex ^scenario-[a-z0-9]{5}$`,
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, t *stepdef.T, name string) (err error) {
		randName := fmt.Sprintf("%s-%s", name, stepdef.RandChars(5))
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randName,
			},
		}
		u := &unstructured.Unstructured{}
		err = t.Client.Scheme().Convert(ns, u, nil)
		if err != nil {
			return err
		}
		err = SetVarFunc(ctx, name, randName)
		if err != nil {
			return err
		}
		return iCreateFunc(ctx, t, u, nil)
	},
}
