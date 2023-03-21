package steps

import (
	"context"
	"time"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	scheme.Default.MustAddToScheme(AsyncAssertExec)
	scheme.Default.MustAddToScheme(AsyncAssertExecWithTimeout)
}

var AsyncAssertExecFunc = func(ctx context.Context, phrase string, timeout time.Duration, ref, not, matcher string, opts *corev1.PodLogOptions) (err error) {
	session := contextutils.LoadSession(ctx, ref)
	Expect(session).ShouldNot(BeNil(), ErrNoResource, ref)

	matcher = "say " + matcher

	NewStringAsyncAssertion(phrase, session).
		WithContext(ctx, timeout).
		ShouldOrShouldNot(not, matcher)

	return nil
}

var AsyncAssertExecWithTimeout = scheme.StepDefinition{
	Name: "it-should-exec-duration",
	Text: "<assertion> <duration> <reference> exec (should|should not) say <text>",
	Help: `Asserts that the referenced pod session has logged something within the specified duration`,
	Examples: `
		Given a resource called sleeping-pod:
		  """
		  apiVersion: v1
		  kind: Pod
		  metadata:
		    name: sleeping-pod
		    namespace: default
		  spec:
		    restartPolicy: Never
		    containers:
		    - command:
		      - sleep
		      - 9000
		      image: busybox:latest
		      name: busybox
		  """
		When I create sleeping-pod
		And I exec "echo hello" in sleeping-pod/busybox
		Then within 30s sleeping-pod exec should say hello`,
	Parameters: []parameters.Parameter{parameters.AsyncAssertionPhrase, parameters.Duration, parameters.Reference, parameters.ShouldOrShouldNot, parameters.Text, parameters.PodLogOptions},
	Function:   AsyncAssertExecFunc,
}

var AsyncAssertExec = scheme.StepDefinition{
	Name: "it-should-exec",
	Text: "<reference> exec (should|should not) say <text>",
	Help: `Asserts that the referenced pod session has logged something`,
	Examples: `
		Given a resource called sleeping-pod:
		  """
		  apiVersion: v1
		  kind: Pod
		  metadata:
		    name: sleeping-pod
		    namespace: default
		  spec:
		    restartPolicy: Never
		    containers:
		    - command:
		      - sleep
		      - 9000
		      image: busybox:latest
		      name: busybox
		  """
		When I create sleeping-pod
		And I exec "echo hello" in sleeping-pod/busybox
		Then sleeping-pod exec should say hello`,
	Parameters: []parameters.Parameter{parameters.Reference, parameters.ShouldOrShouldNot, parameters.Text, parameters.PodLogOptions},
	Function: func(ctx context.Context, ref, not, matcher string, opts *corev1.PodLogOptions) (err error) {
		return AsyncAssertExecFunc(ctx, "", time.Second, ref, not, matcher, opts)
	},
}
