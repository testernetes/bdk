package steps

import (
	"context"

	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	"github.com/testernetes/gkube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IProxyGetFunc = func(ctx context.Context, c gkube.KubernetesHelper, scheme string, ref *unstructured.Unstructured, port, path string, params map[string]string) error {

	switch ref.GetObjectKind().GroupVersionKind().Kind {
	case "Pod":
		pod := &corev1.Pod{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(ref.UnstructuredContent(), pod)
		if err != nil {
			panic(err)
		}
	case "Service":
		service := &corev1.Service{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(ref.UnstructuredContent(), service)
		if err != nil {
			panic(err)
		}
	}

	s, err := c.ProxyGet(ctx, ref, scheme, port, path, params, nil, nil)
	if err != nil {
		return err
	}

	store.Save(ctx, client.ObjectKeyFromObject(ref).String(), s)

	return nil
}

var IProxyGet = stepdef.StepDefinition{
	Name: "i-proxy-get",
	Text: "I proxy get <scheme>://<reference>:<port><path>",
	Help: `Create a proxy connection to the referenced pod resource and attempts a http(s) GET for the port and path.
	Step will fail if the reference was not defined in a previous step.`,
	Examples: `
	Given a resource called pod:
	"""yaml
	apiVersion: v1
	kind: Pod
	metadata:
	  name: app
	  namespace: default
	spec:
	  restartPolicy: Never
	  containers:
	  - command: ["busybox", "httpd", "-f", "-p", "8000"]
	    image: busybox:latest
	    name: server
	"""
	When I create pod
	And within 1m pod jsonpath '{.status.phase}' should equal Running
	And I proxy get http://app:8000/fake
	Then pod response code should equal 404`,
	Function: IProxyGetFunc,
}

var IProxyGetHTTP = stepdef.StepDefinition{
	Function: func(ctx context.Context, c gkube.KubernetesHelper, ref *unstructured.Unstructured, port, path string, options map[string]string) error {
		return IProxyGetFunc(ctx, c, "", ref, port, path, options)
	},
	Name: "i-proxy-get",
	Text: "I proxy get <reference>:<port><path>",
}
