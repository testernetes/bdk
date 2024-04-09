package steps

import (
	"context"
	"time"

	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/gkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AsyncAssertFunc = func(ctx context.Context, c client.WithWatch, assert stepdef.Assert, timeout time.Duration, ref *unstructured.Unstructured, jsonpath string, desiredMatch bool, matcher types.GomegaMatcher) (err error) {
	i, err := c.Watch(ctx, ref)
	if err != nil {
		return err
	}
	defer i.Stop()

	matcher = gkube.HaveJSONPath(jsonpath, matcher)

	retry := true
	for retry {
		select {
		case <-time.After(timeout):
			return
		case <-ctx.Done():
			return
		case event := <-i.ResultChan():
			if event.Type != watch.Modified && event.Type != watch.Added {
				continue
			}

			if event.Type == watch.Error {
				err = errors.New(event.Object.(*metav1.Status).String())
				continue
			}
			retry, err = assert(desiredMatch, matcher, event.Object)
		}
	}
	return nil
}

var AsyncAssertWithTimeout = stepdef.StepDefinition{
	Name: "it-should-object-duration",
	Text: "{assertion} {duration} {reference} jsonpath {jsonpath} {should|should not} {matcher}",
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
	Function: AsyncAssertFunc,
}

var AsyncAssert = stepdef.StepDefinition{
	Name: "it-should-object",
	Text: "{reference} jsonpath {jsonpath} {should|should not} {matcher}",
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
	Function: func(ctx context.Context, c client.WithWatch, ref *unstructured.Unstructured, jsonpath string, desiredMatch bool, matcher types.GomegaMatcher) (err error) {
		return AsyncAssertFunc(ctx, c, stepdef.Eventually, time.Second, ref, jsonpath, desiredMatch, matcher)
	},
}
