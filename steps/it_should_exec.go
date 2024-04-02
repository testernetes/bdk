package steps

import (
	"context"
	"time"

	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/stepdef"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var AsyncAssertExecFunc = func(ctx context.Context, phrase string, timeout time.Duration, ref *unstructured.Unstructured, desiredMatch bool, matcher types.GomegaMatcher, opts *corev1.PodLogOptions) (err error) {

	return nil
}

var AsyncAssertExecWithTimeout = stepdef.StepDefinition{
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
	Function: AsyncAssertExecFunc,
}

var AsyncAssertExec = stepdef.StepDefinition{
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
	Function: func(ctx context.Context, ref *unstructured.Unstructured, desiredMatch bool, matcher types.GomegaMatcher, opts *corev1.PodLogOptions) (err error) {
		return AsyncAssertExecFunc(ctx, "", time.Second, ref, desiredMatch, matcher, opts)
	},
}
