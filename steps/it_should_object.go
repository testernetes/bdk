package steps

import (
	"context"
	"time"

	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/gkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AsyncAssertFunc = func(ctx context.Context, t *stepdef.T, assert stepdef.Assert, timeout time.Duration, ref *unstructured.Unstructured, jsonpath string, desiredMatch bool, matcher types.GomegaMatcher) (err error) {
	i, err := t.Client.Watch(ctx, ref, client.InNamespace(ref.GetNamespace()))
	if err != nil {
		return err
	}
	defer i.Stop()

	matcher = gkube.HaveJSONPath(jsonpath, matcher)
	deadline := time.After(timeout)

	for {
		select {
		case <-deadline:
			t.Log.Info("timedout", "after", timeout)
			return
		case <-ctx.Done():
			return
		case event := <-i.ResultChan():
			if event.Type != watch.Modified && event.Type != watch.Added {
				continue
			}

			if event.Type == watch.Error {
				t.Errorf(event.Object.(*metav1.Status).String())
				continue
			}

			// Watch is always looking at a list
			if event.Object.(client.Object).GetName() != ref.GetName() {
				continue
			}
			retry, err := assert(desiredMatch, matcher, event.Object)
			t.Error(err)
			if !retry {
				return err
			}
		case <-time.Tick(300 * time.Millisecond):
			t.SetProgressGivenDuration(timeout)
		}
	}
}

var AsyncAssertWithTimeout = stepdef.StepDefinition{
	Name: "it-should-object-duration",
	Text: "^{assertion} {duration} {reference} jsonpath {jsonpath} {should|should not} {matcher}$",
	Help: `Asserts that the referenced resource will satisfy the matcher in the specified duration`,
	Examples: `
		Given a resource called cm:
		  """
		  apiVersion: v1
		  kind: ConfigMap
		  metadata:
		    name: example
		    namespace: default
		  """
		And I create cm
		Then within 1s cm jsonpath '{.metadata.uid}' should not be empty`,
	StepArg:  stepdef.NoStepArg,
	Function: AsyncAssertFunc,
}

var AsyncAssert = stepdef.StepDefinition{
	Name: "it-should-object",
	Text: "^{reference} jsonpath {jsonpath} {should|should not} {matcher}$",
	Help: `Asserts that the referenced resource will satisfy the matcher`,
	Examples: `
		Given a resource called cm:
		  """
		  apiVersion: v1
		  kind: ConfigMap
		  metadata:
		    name: example
		    namespace: default
		  """
		And I create cm
		Then cm jsonpath '{.metadata.uid}' should not be empty`,
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, t *stepdef.T, ref *unstructured.Unstructured, jsonpath string, desiredMatch bool, matcher types.GomegaMatcher) (err error) {
		return AsyncAssertFunc(ctx, t, stepdef.Eventually, time.Second, ref, jsonpath, desiredMatch, matcher)
	},
}
