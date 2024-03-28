package steps

import (
	"context"
	"time"

	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/stepdef"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AsyncAssertFunc = func(ctx context.Context, cl client.WithWatch, phrase string, timeout time.Duration, ref *unstructured.Unstructured, jsonpath string, not bool, matcher types.GomegaMatcher) (err error) {

	ctx, _ = context.WithTimeout(ctx, timeout)
	i, err := cl.Watch(ctx, ref)

	for {
		select {
		case <-ctx.Done():
			i.Stop()
			return
		case event := <-i.ResultChan():
			var success bool
			success, err = matcher.Match(event.Object)
			if err != nil {
				continue
			}
			if success {
				return nil
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
