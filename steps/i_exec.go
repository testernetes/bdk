package steps

import (
	"context"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	"github.com/testernetes/gkube"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IExecFunc = func(ctx context.Context, c gkube.KubernetesHelper, cmd []string, pod *corev1.Pod, container string) error {
	return clientDo(ctx, func() error {
		s, err := c.Exec(ctx, pod, container, cmd, nil, nil)
		if err != nil {
			return err
		}
		store.Save(ctx, client.ObjectKeyFromObject(pod).String(), s)
		return nil
	})
}

var IExecScriptFunc = func(ctx context.Context, c gkube.KubernetesHelper, pod *corev1.Pod, container string, script *messages.DocString) error {
	shell := script.MediaType
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := script.Content
	return IExecFunc(ctx, c, []string{shell, "-c", cmd}, pod, container)
}

var IExecInContainer = stepdef.StepDefinition{
	Name: "i-exec-in-container",
	Text: "I exec <command> in <reference>/<container>",
	Help: "Executes the given command in a shell in the referenced pod and container.",
	Examples: `
	When I exec "echo helloworld" in pod/app`,
	Function: func(ctx context.Context, c gkube.KubernetesHelper, cmd string, pod *corev1.Pod, container string) error {
		return IExecFunc(ctx, c, []string{"/bin/sh", "-c", cmd}, pod, container)
	},
}

var IExecInDefaultContainer = stepdef.StepDefinition{
	Name: "i-exec",
	Text: "I exec <command> in <reference>",
	Help: "Executes the given command in a shell in the referenced pod and default container.",
	Examples: `
	When I exec "echo helloworld" in pod`,
	Function: func(ctx context.Context, c gkube.KubernetesHelper, cmd string, pod *corev1.Pod) error {
		return IExecFunc(ctx, c, []string{"/bin/sh", "-c", cmd}, pod, "")
	},
}

var IExecScriptInContainer = stepdef.StepDefinition{
	Name: "i-exec-script-in-container",
	Text: "I exec this script in <reference>/<container>",
	Help: "Executes the given script in a shell in the referenced pod and container.",
	Examples: `
	When I exec this script in pod/app
	  """/bin/bash
	  curl localhost:8080/ready
	  """`,
	StepArg:  stepdef.Script,
	Function: IExecScriptFunc,
}

var IExecScriptInDefaultContainer = stepdef.StepDefinition{
	Name: "i-exec-script",
	Text: "I exec this script in <reference>",
	Help: "Executes the given script in a shell in the referenced pod and default container.",
	Examples: `
	When I exec this script in pod
	  """/bin/bash
	  curl localhost:8080/ready
	  """`,
	Function: func(ctx context.Context, c gkube.KubernetesHelper, pod *corev1.Pod, script *messages.DocString) error {
		return IExecScriptFunc(ctx, c, pod, "", script)
	},
}
