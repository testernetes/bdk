package steps

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/testernetes/bdk/stepdef"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var AsyncAssertLogFunc = func(ctx context.Context, c kubernetes.Clientset, assert stepdef.Assert, timeout time.Duration, pod *corev1.Pod, desiredMatch bool, text string, opts *corev1.PodLogOptions) (err error) {
	stream, err := c.CoreV1().
		Pods(pod.Namespace).
		GetLogs(pod.Name, opts).
		Stream(ctx)
	if err != nil {
		return err
	}

	s := &PodSession{
		Out: gbytes.NewBuffer(),
	}

	go func() {
		defer stream.Close()
		_, err = io.Copy(s.Out, stream)
	}()

	defer s.Out.CancelDetects()

	retry := true
	for retry {
		select {
		case <-time.After(timeout):
			if desiredMatch {
				return fmt.Errorf("did not find '%s' in logs:\n%s", text, s.Out.Contents())
			}
		case <-ctx.Done():
			return
		case <-s.Out.Detect(text):
			if desiredMatch {
				return nil
			}
			return fmt.Errorf("found '%s' in logs:\n%s", text, s.Out.Contents())
		}
	}
	return nil
}

var AsyncAssertLogWithTimeout = stepdef.StepDefinition{
	Name: "it-should-log-duration",
	Text: "{assertion} {duration} {reference} logs {should|should not} say {text}",
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
	StepArg:  stepdef.PodLogOptions,
	Function: AsyncAssertLogFunc,
}

var AsyncAssertLog = stepdef.StepDefinition{
	Name: "it-should-log",
	Text: "{reference} logs {should|should not} say {text}",
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
	StepArg: stepdef.PodLogOptions,
	Function: func(ctx context.Context, c kubernetes.Clientset, pod *corev1.Pod, desiredMatch bool, text string, opts *corev1.PodLogOptions) (err error) {
		return AsyncAssertLogFunc(ctx, c, stepdef.Eventually, time.Second, pod, desiredMatch, text, opts)
	},
}

// TODO maybe parse the multiline from comments
// TODO maybe set the pod log options in a previous step
// TODO maybe embed podLogOptions in DocString
//var AsyncAssertLogWithTimeoutMultiline = scheme.StepDefinition{
//	Name: "it-should-log-duration-multiline",
//	Text: "{assertion} {duration} {reference} logs (should|should not) say",
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
