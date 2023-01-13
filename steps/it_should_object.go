package steps

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/register"
	"github.com/testernetes/bdk/scheme"
)

func init() {
	scheme.Default.MustAddToScheme(AsyncAssert)
	scheme.Default.MustAddToScheme(AsyncAssertWithTimeout)
}

var AsyncAssertFunc = func(ctx context.Context, phrase, timeout, ref, jsonpath, not, matcher string) (err error) {
	u := register.Load(ctx, ref)
	Expect(u).ShouldNot(BeNil(), ErrNoResource, ref)

	// nest the jsonpath transformer with the matcher
	matcher = fmt.Sprintf("jsonpath '%s' %s", jsonpath, matcher)

	c := client.MustGetClientFrom(ctx)
	NewStringAsyncAssertion(phrase, c.Object).
		WithContext(ctx, timeout).
		WithArguments(u).
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
	Parameters: []parameters.Parameter{parameters.AsyncAssertion, parameters.Duration, parameters.Reference, parameters.JSONPath, parameters.ShouldOrShouldNot, parameters.Matcher},
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
		return AsyncAssertFunc(ctx, "", "", ref, jsonpath, not, matcher)
	},
}
