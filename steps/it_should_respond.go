package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/gkube"
)

var AsyncAssertRespFunc = func(ctx context.Context, assert stepdef.Assert, timeout time.Duration, s gkube.PodSession, desiredMatch bool, text string) (err error) {
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
		case <-s.Buffer().Detect(text):
			if desiredMatch {
				return nil
			}
			return fmt.Errorf("found '%s' in logs:\n%s", text, s.Out.Contents())
		}
	}

	return nil
}

var AsyncAssertRespWithTimeout = stepdef.StepDefinition{
	Name: "it-should-resp-duration",
	Text: "^{assertion} {duration} {reference} response {should|should not} say {text}$",
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
	StepArg:  stepdef.NoStepArg,
	Function: AsyncAssertRespFunc,
}

var AsyncAssertResp = stepdef.StepDefinition{
	Name: "it-should-resp",
	Text: "^{reference} response {should|should not} say {text}$",
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
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, ref gkube.PodSession, desiredMatch bool, text string) (err error) {
		return AsyncAssertRespFunc(ctx, stepdef.Eventually, time.Second, ref, desiredMatch, text)
	},
}
