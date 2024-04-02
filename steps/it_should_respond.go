package steps

import (
	"context"
	"time"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/stepdef"
)

var AsyncAssertRespFunc = func(ctx context.Context, phrase string, timeout time.Duration, ref, not, matcher string) (err error) {
	session := contextutils.LoadSession(ctx, ref)
	Expect(session).ShouldNot(BeNil(), ErrNoResource, ref)

	matcher = "say " + matcher

	NewStringAsyncAssertion(phrase, session).
		WithContext(ctx, timeout).
		ShouldOrShouldNot(not, matcher)

	return nil
}

var AsyncAssertRespWithTimeout = stepdef.StepDefinition{
	Name: "it-should-resp-duration",
	Text: "<assertion> <duration> <reference> response (should|should not) say <text>",
	Help: `Asserts that the referenced pod session has responded with something within the specified duration`,
	Examples: `
		Given a resource called sleeping-pod:
		  """
		  apiVersion: v1
		  kind: Pod
		  metadata:
		    name: my-api
		    namespace: default
		  spec:
		    restartPolicy: Never
		    containers:
		    - command:
		      - nc
		      - create a real example
		      image: busybox:latest
		      name: busybox
		  """
		When I create my-api
		And I proxy get http://my-app:8000/fake
		Then within 30s my-api response should say hello`,
	Function: AsyncAssertRespFunc,
}

var AsyncAssertResp = stepdef.StepDefinition{
	Name: "it-should-resp",
	Text: "<reference> response (should|should not) say <text>",
	Help: `Asserts that the referenced pod session has logged something`,
	Examples: `
		Given a resource called sleeping-pod:
		  """
		  apiVersion: v1
		  kind: Pod
		  metadata:
		    name: my-api
		    namespace: default
		  spec:
		    restartPolicy: Never
		    containers:
		    - command:
		      - nc
		      - create a real example
		      image: busybox:latest
		      name: busybox
		  """
		When I create my-api
		And I proxy get http://my-app:8000/fake
		Then within 30s my-api response should say hello`,
	Function: func(ctx context.Context, ref, not, matcher string) (err error) {
		return AsyncAssertRespFunc(ctx, "", time.Second, ref, not, matcher)
	},
}
