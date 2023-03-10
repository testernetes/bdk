package steps

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
)

func init() {
	scheme.Default.MustAddToScheme(AsyncAssert)
	scheme.Default.MustAddToScheme(AsyncAssertWithTimeout)
}

var AsyncAssertFunc = func(ctx context.Context, phrase string, timeout time.Duration, ref, jsonpath, not, matcher string) (err error) {
	o := contextutils.LoadObject(ctx, ref)
	Expect(o).ShouldNot(BeNil(), ErrNoResource, ref)

	// nest the jsonpath transformer with the matcher
	matcher = fmt.Sprintf("jsonpath '%s' %s", jsonpath, matcher)

	c := contextutils.MustGetClientFrom(ctx)
	NewStringAsyncAssertion(phrase, c.Object).
		WithContext(ctx, timeout).
		WithArguments(o).
		ShouldOrShouldNot(not, matcher)

	return nil
}

var AsyncAssertWithTimeout = scheme.StepDefinition{
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
	Parameters: []parameters.Parameter{parameters.AsyncAssertionPhrase, parameters.Duration, parameters.Reference, parameters.JSONPath, parameters.ShouldOrShouldNot, parameters.Matcher},
	Function:   AsyncAssertFunc,
}

var AsyncAssert = scheme.StepDefinition{
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
	Parameters: []parameters.Parameter{parameters.Reference, parameters.JSONPath, parameters.ShouldOrShouldNot, parameters.Matcher},
	Function: func(ctx context.Context, ref, jsonpath, not, matcher string) (err error) {
		return AsyncAssertFunc(ctx, "", time.Second, ref, jsonpath, not, matcher)
	},
}
