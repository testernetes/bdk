package steps

import (
	"context"
	"fmt"
	"io"

	"github.com/onsi/gomega/gbytes"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var IProxyGetFunc = func(ctx context.Context, t *stepdef.T, scheme string, obj *unstructured.Unstructured, port, path string, params map[string]string) (err error) {
	session := &PodSession{
		Out: gbytes.NewBuffer(),
	}
	var stream io.ReadCloser

	switch obj.GetObjectKind().GroupVersionKind().Kind {
	case "Pod":
		stream, err = t.Clientset.CoreV1().
			Pods(obj.GetNamespace()).ProxyGet(scheme, obj.GetName(), port, path, params).
			Stream(ctx)
	case "Service":
		stream, err = t.Clientset.CoreV1().
			Services(obj.GetNamespace()).ProxyGet(scheme, obj.GetName(), port, path, params).
			Stream(ctx)
	default:
		return fmt.Errorf("expected a Pod or Service, got %T", obj)
	}
	if err != nil {
		return err
	}

	_, err = io.Copy(session.Out, stream)

	store.Save(ctx, client.ObjectKeyFromObject(obj).String(), session)

	return nil
}

var IProxyGet = stepdef.StepDefinition{
	Name: "i-proxy-get",
	Text: "^I proxy get {scheme}://{reference}:{port}{path}$",
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
	StepArg:  stepdef.ProxyGetOptions,
	Function: IProxyGetFunc,
}

var IProxyGetHTTP = stepdef.StepDefinition{
	Name:    "i-proxy-get",
	Text:    "^I proxy get {reference}:{port}{path}$",
	StepArg: stepdef.ProxyGetOptions,
	Function: func(ctx context.Context, t *stepdef.T, ref *unstructured.Unstructured, port, path string, options map[string]string) error {
		return IProxyGetFunc(ctx, t, "", ref, port, path, options)
	},
}
