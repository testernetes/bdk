package steps

import (
	"context"
	"fmt"
	"regexp"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/models"
	"github.com/testernetes/bdk/register"
)

func init() {
	err := models.Scheme.Register(AsyncAssertWithTimeout)
	if err != nil {
		panic(err)
	}
	err = models.Scheme.Register(AsyncAssert)
	if err != nil {
		panic(err)
	}
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

var AsyncAssertWithTimeout = models.StepDefinition{
	Expression: regexp.MustCompile(fmt.Sprintf("^%s %s %s jsonpath %s %s %s$", AsyncType, Duration, NamedObj, SingleQuoted, ShouldOrShouldNot, Matcher)),
	Function:   AsyncAssertFunc,
	Name:       "it-should-object",
	Text:       "<assertion> <duration> <reference> (should|should not) <matcher>",
}

var AsyncAssert = models.StepDefinition{
	Expression: regexp.MustCompile(fmt.Sprintf("^%s jsonpath %s %s %s$", NamedObj, SingleQuoted, ShouldOrShouldNot, Matcher)),
	Function: func(ctx context.Context, ref, jsonpath, not, matcher string) (err error) {
		return AsyncAssertFunc(ctx, "", "", ref, jsonpath, not, matcher)
	},
	Name: "it-should-object",
	Text: "<reference> (should|should not) <matcher>",
}
