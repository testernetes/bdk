package steps

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/arguments"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/register"
	"github.com/testernetes/bdk/scheme"
	"github.com/testernetes/gkube"
)

func init() {
	scheme.Default.MustAddToScheme(AsyncAssertLog)
	scheme.Default.MustAddToScheme(AsyncAssertLogWithTimeout)
}

var AsyncAssertLogFunc = func(ctx context.Context, phrase, timeout, ref, not, matcher string, opts *arguments.DataTable) (err error) {
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

	matcher = "say " + matcher

	NewStringAsyncAssertion(phrase, s).
		WithContext(ctx, timeout).
		ShouldOrShouldNot(not, matcher)

	return nil
}

var AsyncAssertLogWithTimeout = scheme.StepDefinition{
	Name: "it-should-log-duration",
	Text: "<assertion> <duration> <reference> logs (should|should not) say <text>",
	Help: `Asserts that the referenced resource will log something within the specified duration`,
	Examples: `
		Given a resource called testernetes:
		  """
		  apiVersion: v1
		  kind: Pod
		  metadata:
		    name: testernetes
		    namespace: default
		  spec:
		    restartPolicy: Never
		    containers:
		    - command:
		      - /bdk
		      - --help
		      image: ghcr.io/testernetes/bdk:d408c829f019f2052badcb93282ee6bd3cdaf8d0
		      name: bdk
		  """
		When I create testernetes
		Then within 1m testernetes logs should say Behaviour Driven Kubernetes
		  | container | bdk   |
		  | follow    | false |`,
	Parameters: []parameters.Parameter{parameters.AsyncAssertion, parameters.Duration, parameters.Reference, parameters.ShouldOrShouldNot, parameters.Text, parameters.PodLogOptions},
	Function:   AsyncAssertLogFunc,
}

var AsyncAssertLog = scheme.StepDefinition{
	Name: "it-should-log",
	Text: "<reference> logs (should|should not) say <text>",
	Help: `Asserts that the referenced resource will log something within the specified duration`,
	Examples: `
		Given a resource called testernetes:
		  """
		  apiVersion: v1
		  kind: Pod
		  metadata:
		    name: testernetes
		    namespace: default
		  spec:
		    restartPolicy: Never
		    containers:
		    - command:
		      - /bdk
		      - --help
		      image: ghcr.io/testernetes/bdk:d408c829f019f2052badcb93282ee6bd3cdaf8d0
		      name: bdk
		  """
		When I create testernetes
		Then testernetes logs should say Behaviour Driven Kubernetes`,
	Parameters: []parameters.Parameter{parameters.Reference, parameters.ShouldOrShouldNot, parameters.Text, parameters.PodLogOptions},
	Function: func(ctx context.Context, ref, not, matcher string, opts *arguments.DataTable) (err error) {
		return AsyncAssertLogFunc(ctx, "", "", ref, not, matcher, opts)
	},
}

// TODO maybe parse the multiline from comments
// TODO maybe set the pod log options in a previous step
// TODO maybe embed podLogOptions in DocString
//var AsyncAssertLogWithTimeoutMultiline = scheme.StepDefinition{
//	Name: "it-should-log-duration-multiline",
//	Text: "<assertion> <duration> <reference> logs (should|should not) say",
//	Help: `Assets that the referenced resource will log something within the specified duration`,
//	Examples: `
//		Given a resource called bdk:
//		  """
//		  apiVersion: v1
//		  kind: Pod
//		  metadata:
//		    name: bdk
//		    namespace: default
//		  spec:
//		    restartPolicy: Never
//		    containers:
//		    - command:
//		      - /bdk
//		      - --help
//		      image: ghcr.io/testernetes/bdk:d408c829f019f2052badcb93282ee6bd3cdaf8d0
//		      name: bdk
//		  """
//		When I create bdk
//		Then within 1m bdk logs should say
//		  """
//		  Behaviour Driven Kubernetes
//
//		  Usage:
//		    bdk [command]
//		  """`,
//	Parameters: []parameters.Parameter{parameters.AsyncAssertion, parameters.Duration, parameters.Reference, parameters.ShouldOrShouldNot, parameters.MultilineText},
//	Function: func(ctx context.Context, ref, not, matcher string, opts *messages.DataTable) (err error) {
//		return AsyncAssertLogFunc(ctx, "", "", ref, not, matcher, opts)
//	},
//}
