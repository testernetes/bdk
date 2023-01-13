package steps

import (
	"context"
	"fmt"
	"regexp"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/models"
	"github.com/testernetes/bdk/register"
	"github.com/testernetes/gkube"
)

func init() {
	err := models.Scheme.Register(AsyncAssertLog)
	if err != nil {
		panic(err)
	}
	err = models.Scheme.Register(AsyncAssertLogWithTimeout)
	if err != nil {
		panic(err)
	}
}

var AsyncAssertLogFunc = func(ctx context.Context, phrase, timeout, ref, not, matcher string, opts *messages.DataTable) (err error) {
	pod := register.LoadPod(ctx, ref)
	Expect(pod).ShouldNot(BeNil(), ErrNoResource, ref)

	//out, errOut := writer.From(ctx)

	podLogOptions := client.PodLogOptionsFrom(opts)

	var s *gkube.PodSession
	c := client.MustGetClientFrom(ctx)
	NewStringAsyncAssertion("", func() error {
		s, err = c.Logs(ctx, pod, podLogOptions, nil, nil)
		return err
	}).WithContext(ctx, timeout).Should(Succeed())

	NewStringAsyncAssertion(phrase, s).
		WithContext(ctx, timeout).
		ShouldOrShouldNot(not, matcher)

	return nil
}

var AsyncAssertLogWithTimeout = models.StepDefinition{
	Name:       "it-should-log",
	Text:       "<assertion> <duration> <reference> logs (should|should not) say <string>",
	Expression: regexp.MustCompile(fmt.Sprintf("^%s %s %s logs %s %s$", AsyncType, Duration, NamedObj, ShouldOrShouldNot, Matcher)),
	Function:   AsyncAssertLogFunc,
}

var AsyncAssertLog = models.StepDefinition{
	Name:       "it-should-log",
	Text:       "<reference> logs (should|should not) say <string>",
	Expression: regexp.MustCompile(fmt.Sprintf("^%s logs %s %s$", NamedObj, ShouldOrShouldNot, Matcher)),
	Function: func(ctx context.Context, ref, not, matcher string, opts *messages.DataTable) (err error) {
		return AsyncAssertLogFunc(ctx, "", "", ref, not, matcher, opts)
	},
}
