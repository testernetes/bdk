package steps

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/models"
	"github.com/testernetes/bdk/register"
	"github.com/testernetes/bdk/sessions"
	"github.com/testernetes/gkube"
)

func init() {
	err := models.Scheme.Register(IExecInContainer)
	if err != nil {
		panic(err)
	}
	err = models.Scheme.Register(IExecInDefaultContainer)
	if err != nil {
		panic(err)
	}
	err = models.Scheme.Register(IExecScriptInContainer)
	if err != nil {
		panic(err)
	}
	err = models.Scheme.Register(IExecScriptInDefaultContainer)
	if err != nil {
		panic(err)
	}
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

	sessions.Save(ctx, ref, s)

	return nil
}

var IExecScriptFunc = func(ctx context.Context, ref, container string, script *models.DocString) error {
	shell := script.MediaType
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := script.Content
	return IExecFunc(ctx, []string{shell, "-c", cmd}, ref, container)
}

var IExecInContainer = models.StepDefinition{
	Name: "i-exec-in-container",
	Text: "I exec <command> in <reference>/<container>",
	Help: "Executes the given command in a shell in the referenced pod and container.",
	Examples: `
	When I exec "echo helloworld" in pod/app`,
	Parameters: []models.Parameter{models.Command, models.Reference, models.Container},
	Function: func(ctx context.Context, cmd string, ref, container string) error {
		return IExecFunc(ctx, []string{"/bin/sh", "-c", cmd}, ref, container)
	},
}

var IExecInDefaultContainer = models.StepDefinition{
	Name: "i-exec",
	Text: "I exec <command> in <reference>",
	Help: "Executes the given command in a shell in the referenced pod and default container.",
	Examples: `
	When I exec "echo helloworld" in pod`,
	Parameters: []models.Parameter{models.Command, models.Reference},
	Function: func(ctx context.Context, cmd, ref string) error {
		return IExecFunc(ctx, []string{"/bin/sh", "-c", cmd}, ref, "")
	},
}

var IExecScriptInContainer = models.StepDefinition{
	Name: "i-exec-script-in-container",
	Text: "I exec this script in <reference>/<container>",
	Help: "Executes the given script in a shell in the referenced pod and container.",
	Examples: `
	When I exec this script in pod/app
	  """/bin/bash
	  curl localhost:8080/ready
	  """`,
	Parameters: []models.Parameter{models.Command, models.Reference, models.Container, models.Script},
	Function:   IExecScriptFunc,
}

var IExecScriptInDefaultContainer = models.StepDefinition{
	Name: "i-exec-script",
	Text: "I exec this script in <reference>",
	Help: "Executes the given script in a shell in the referenced pod and default container.",
	Examples: `
	When I exec this script in pod
	  """/bin/bash
	  curl localhost:8080/ready
	  """`,
	Parameters: []models.Parameter{models.Command, models.Reference, models.Script},
	Function: func(ctx context.Context, ref string, script *models.DocString) error {
		return IExecScriptFunc(ctx, ref, "", script)
	},
}
