package steps

import (
	"context"
	"fmt"
	"regexp"
	"time"

	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
	"github.com/testernetes/gkube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	scheme.Default.MustAddToScheme(IProxyGet)
}

var IProxyGetFunc = func(ctx context.Context, scheme, ref, port, path string, params map[string]string) error {
	u := contextutils.LoadObject(ctx, ref)
	Expect(u).ShouldNot(BeNil(), ErrNoResource, ref)

	var o client.Object
	switch u.GetObjectKind().GroupVersionKind().Kind {
	case "Pod":
		pod := &corev1.Pod{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), pod)
		if err != nil {
			panic(err)
		}
		o = pod
	case "Service":
		service := &corev1.Service{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), service)
		if err != nil {
			panic(err)
		}
		o = service
	}

	var s *gkube.PodSession
	c := contextutils.MustGetClientFrom(ctx)
	Eventually(func() error {
		var err error
		s, err = c.ProxyGet(ctx, o, scheme, port, path, params, nil, nil)
		return err
	}).WithTimeout(time.Minute).Should(Succeed(), "Failed to proxy get")

	_ = s

	return nil
}

var IProxyGet = scheme.StepDefinition{
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
	Parameters: []parameters.Parameter{parameters.URLScheme, parameters.Reference, parameters.Port, parameters.URLPath, parameters.ProxyGetOptions},
	Function:   IProxyGetFunc,
}

var IProxyGetHTTP = scheme.StepDefinition{
	Expression: regexp.MustCompile(fmt.Sprintf(`^I proxy get %s%s%s$`, NamedObj, Port, URLPath)),
	Function: func(ctx context.Context, ref, port, path string, options map[string]string) error {
		return IProxyGetFunc(ctx, "", ref, port, path, options)
	},
	Name: "i-proxy-get",
	Text: "I proxy get <reference>:<port><path>",
}
