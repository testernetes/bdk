package steps

import (
	"context"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var AsyncAssertExecFunc = func(ctx context.Context, assert stepdef.Assert, timeout time.Duration, ref *unstructured.Unstructured, desiredMatch bool, text string) (err error) {

	session := store.Load[PodSession](ctx, client.ObjectKeyFromObject(ref).String())
	matcher := gbytes.Say(text)

	_, err = assert(desiredMatch, matcher, session.Out)
	if err != nil {
		return err
	}

	return nil
}

var AsyncAssertExecWithTimeout = stepdef.StepDefinition{
	Name: "it-should-exec-duration",
	Text: "^{assertion} {duration} {reference} exec {should|should not} say {text}$",
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
	StepArg:  stepdef.NoStepArg,
	Function: AsyncAssertExecFunc,
}

var AsyncAssertExec = stepdef.StepDefinition{
	Name: "it-should-exec",
	Text: "^{reference} exec {should|should not} say {text}$",
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
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, ref *unstructured.Unstructured, desiredMatch bool, matcher string) (err error) {
		return AsyncAssertExecFunc(ctx, stepdef.Eventually, time.Second, ref, desiredMatch, matcher)
	},
}
