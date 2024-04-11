package steps

import (
	"context"
	"fmt"
	"net/http"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/onsi/gomega/gbytes"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	k8sExec "k8s.io/utils/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type PodSession struct {
	Out      *gbytes.Buffer
	Err      *gbytes.Buffer
	ExitCode int
}

var IExecFunc = func(ctx context.Context, c kubernetes.Clientset, cmd []string, pod *corev1.Pod, container string) error {
	podRestInterface := c.CoreV1().RESTClient()
	//podRestInterface, err := apiutil.RESTClientForGVK(pod.GroupVersionKind(), false, config.GetConfigOrDie(), serializer.NewCodecFactory(c.Scheme()))

	podExecOpts := &corev1.PodExecOptions{
		Stdout:    true,
		Stderr:    true,
		Stdin:     false,
		TTY:       false,
		Container: container,
		Command:   cmd,
	}

	execReq := podRestInterface.Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		VersionedParams(podExecOpts, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config.GetConfigOrDie(), http.MethodPost, execReq.URL())
	if err != nil {
		return err
	}

	session := &PodSession{
		Out: &gbytes.Buffer{},
		Err: &gbytes.Buffer{},
	}
	defer session.Out.Close()
	defer session.Err.Close()

	streamOpts := remotecommand.StreamOptions{
		Stdout: session.Out,
		Stderr: session.Err,
	}
	err = executor.StreamWithContext(ctx, streamOpts)
	if err != nil {
		fmt.Fprintf(session.Err, err.Error())
		if exitcode, ok := err.(k8sExec.CodeExitError); ok {
			session.ExitCode = exitcode.Code
		}
	}
	store.Save(ctx, client.ObjectKeyFromObject(pod).String(), session)
	return err
}

var IExecScriptFunc = func(ctx context.Context, c kubernetes.Clientset, pod *corev1.Pod, container string, script *messages.DocString) error {
	shell := script.MediaType
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := script.Content
	return IExecFunc(ctx, c, []string{shell, "-c", cmd}, pod, container)
}

var IExecInContainer = stepdef.StepDefinition{
	Name: "i-exec-in-container",
	Text: "I exec {command} in {reference}/{container}",
	Help: "Executes the given command in a shell in the referenced pod and container.",
	Examples: `
	When I exec "echo helloworld" in pod/app`,
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, c kubernetes.Clientset, cmd string, pod *corev1.Pod, container string) error {
		return IExecFunc(ctx, c, []string{"/bin/sh", "-c", cmd}, pod, container)
	},
}

var IExecInDefaultContainer = stepdef.StepDefinition{
	Name: "i-exec",
	Text: "I exec {command} in {reference}",
	Help: "Executes the given command in a shell in the referenced pod and default container.",
	Examples: `
	When I exec "echo helloworld" in pod`,
	StepArg: stepdef.NoStepArg,
	Function: func(ctx context.Context, c kubernetes.Clientset, cmd string, pod *corev1.Pod) error {
		return IExecFunc(ctx, c, []string{"/bin/sh", "-c", cmd}, pod, "")
	},
}

var IExecScriptInContainer = stepdef.StepDefinition{
	Name: "i-exec-script-in-container",
	Text: "I exec this script in {reference}/{container}",
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
	Text: "I exec this script in {reference}",
	Help: "Executes the given script in a shell in the referenced pod and default container.",
	Examples: `
	When I exec this script in pod
	  """/bin/bash
	  curl localhost:8080/ready
	  """`,
	StepArg: stepdef.Script,
	Function: func(ctx context.Context, c kubernetes.Clientset, pod *corev1.Pod, script *messages.DocString) error {
		return IExecScriptFunc(ctx, c, pod, "", script)
	},
}
