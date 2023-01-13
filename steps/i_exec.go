package steps

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/arguments"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/register"
	"github.com/testernetes/bdk/scheme"
	"github.com/testernetes/bdk/session"
	"github.com/testernetes/gkube"
)

func init() {
	scheme.Default.AddToScheme(IExecInContainer)
	scheme.Default.AddToScheme(IExecInDefaultContainer)
	scheme.Default.AddToScheme(IExecScriptInContainer)
	scheme.Default.AddToScheme(IExecScriptInDefaultContainer)
}

var IExecFunc = func(ctx context.Context, cmd []string, ref, container string) error {
	pod := register.LoadPod(ctx, ref)
	Expect(pod).ShouldNot(BeNil(), ErrNoResource, ref)

	//out, errOut := writer.From(ctx)

	var s *gkube.PodSession
	c := client.MustGetClientFrom(ctx)
	Eventually(func() error {
		var err error
		s, err = c.Exec(ctx, pod, container, cmd, nil, nil)
		return err
	}).WithContext(ctx).Should(Succeed(), "Could not exec in container")

	session.Save(ctx, ref, s)

	return nil
}

var IExecScriptFunc = func(ctx context.Context, ref, container string, script *arguments.DocString) error {
	shell := script.MediaType
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := script.Content
	return IExecFunc(ctx, []string{shell, "-c", cmd}, ref, container)
}

var IExecInContainer = scheme.StepDefinition{
	Name: "i-exec-in-container",
	Text: "I exec <command> in <reference>/<container>",
	Help: "Executes the given command in a shell in the referenced pod and container.",
	Examples: `
	When I exec "echo helloworld" in pod/app`,
	Parameters: []parameters.Parameter{parameters.Command, parameters.Reference, parameters.Container},
	Function: func(ctx context.Context, cmd string, ref, container string) error {
		return IExecFunc(ctx, []string{"/bin/sh", "-c", cmd}, ref, container)
	},
}

var IExecInDefaultContainer = scheme.StepDefinition{
	Name: "i-exec",
	Text: "I exec <command> in <reference>",
	Help: "Executes the given command in a shell in the referenced pod and default container.",
	Examples: `
	When I exec "echo helloworld" in pod`,
	Parameters: []parameters.Parameter{parameters.Command, parameters.Reference},
	Function: func(ctx context.Context, cmd, ref string) error {
		return IExecFunc(ctx, []string{"/bin/sh", "-c", cmd}, ref, "")
	},
}

var IExecScriptInContainer = scheme.StepDefinition{
	Name: "i-exec-script-in-container",
	Text: "I exec this script in <reference>/<container>",
	Help: "Executes the given script in a shell in the referenced pod and container.",
	Examples: `
	When I exec this script in pod/app
	  """/bin/bash
	  curl localhost:8080/ready
	  """`,
	Parameters: []parameters.Parameter{parameters.Command, parameters.Reference, parameters.Container, parameters.Script},
	Function:   IExecScriptFunc,
}

var IExecScriptInDefaultContainer = scheme.StepDefinition{
	Name: "i-exec-script",
	Text: "I exec this script in <reference>",
	Help: "Executes the given script in a shell in the referenced pod and default container.",
	Examples: `
	When I exec this script in pod
	  """/bin/bash
	  curl localhost:8080/ready
	  """`,
	Parameters: []parameters.Parameter{parameters.Command, parameters.Reference, parameters.Script},
	Function: func(ctx context.Context, ref string, script *arguments.DocString) error {
		return IExecScriptFunc(ctx, ref, "", script)
	},
}
