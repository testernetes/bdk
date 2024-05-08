package steps

import (
	"context"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/stepdef"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AsyncExistFunc = func(ctx context.Context, t *stepdef.T, assert stepdef.Assert, timeout time.Duration, ref *unstructured.Unstructured, desiredMatch bool) (err error) {
	not := !desiredMatch

	deadline := time.After(timeout)

	err = t.Client.Get(ctx, client.ObjectKeyFromObject(ref), ref)
	if k8sErrors.IsNotFound(err) && !not {
		assert(not, gomega.BeNil(), ref)
	}

	i, err := t.Client.Watch(ctx, ref, client.InNamespace(ref.GetNamespace()))
	if err != nil {
		return err
	}
	defer i.Stop()

	err = k8sErrors.NewNotFound(schema.GroupResource{Group: ref.GetObjectKind().GroupVersionKind().Group, Resource: ref.GetKind()}, ref.GetName())
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
			var retry bool
			retry, err = assert(not, gomega.BeNil(), event.Object)
			if !retry {
				return err
			}
		case <-time.Tick(300 * time.Millisecond):
			t.Error(err)
			t.SetProgressGivenDuration(timeout)
		}
	}
}

var AsyncExistWithTimeout = stepdef.StepDefinition{
	Name: "it-should-exist-duration",
	Text: "^{assertion} {duration} {reference} {should|should not} exist$",
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
		Then within 1s cm should exist`,
	StepArg:  stepdef.NoStepArg,
	Function: AsyncExistFunc,
}

var AsyncExist = stepdef.StepDefinition{
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
		Then cm should exist`,
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, t *stepdef.T, ref *unstructured.Unstructured, jsonpath string, desiredMatch bool, matcher types.GomegaMatcher) (err error) {
		return AsyncExistFunc(ctx, t, stepdef.Eventually, time.Second, ref, desiredMatch)
	},
}
