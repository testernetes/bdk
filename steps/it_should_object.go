package steps

import (
	"context"
	"time"

	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"github.com/testernetes/bdk/stepdef"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AsyncAssertFunc = func(ctx context.Context, cl client.WithWatch, assertion stepdef.Assertion, timeout time.Duration, ref *unstructured.Unstructured, desiredMatch bool, matcher types.GomegaMatcher) (err error) {

	ctx, _ = context.WithTimeout(ctx, timeout)

	i, err := cl.Watch(ctx, ref)
	if err != nil {
		return err
	}
	defer i.Stop()

	for {
		select {
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

			var matches bool
			matches, err = matcher.Match(event.Object)
			if err != nil {
				return err
			}

			if matches == desiredMatch {
				if assertion == stepdef.EventuallyAssertion {
					return nil
				}
			}

			if matches != desiredMatch {
				if desiredMatch {
					err = errors.New(matcher.FailureMessage(event.Object))
				} else {
					err = errors.New(matcher.NegatedFailureMessage(event.Object))
				}
				if assertion == stepdef.ConsistentlyAssertion {
					return err
				}
			}
		}
	}
	return nil
}

var AsyncAssertWithTimeout = stepdef.StepDefinition{
	Name: "it-should-object-duration",
	Text: "<assertion> <duration> <reference> jsonpath <jsonpath> (should|should not) <matcher>",
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
	Text: "<reference> jsonpath <jsonpath> (should|should not) <matcher>",
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
	Function: func(ctx context.Context, ref, jsonpath, not, matcher string) (err error) {
		return AsyncAssertFunc(ctx, "", time.Second, ref, jsonpath, not, matcher)
	},
}
