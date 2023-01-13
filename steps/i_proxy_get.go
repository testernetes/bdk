package steps

import (
	"context"
	"fmt"
	"regexp"
	"time"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/client"
	"github.com/testernetes/bdk/models"
	"github.com/testernetes/bdk/register"
	"github.com/testernetes/gkube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	err := models.Scheme.Register(IProxyGet)
	if err != nil {
		panic(err)
	}
}

var IProxyGetFunc = func(ctx context.Context, scheme, ref, port, path string, options *messages.DataTable) error {
	u := register.Load(ctx, ref)
	Expect(u).ShouldNot(BeNil(), ErrNoResource, ref)

	var o ctrlclient.Object
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
	params := client.ProxyGetOptionsFrom(options)

	var s *gkube.PodSession
	c := client.MustGetClientFrom(ctx)
	Eventually(func() error {
		var err error
		s, err = c.ProxyGet(ctx, o, scheme, port, path, params, nil, nil)
		return err
	}).WithTimeout(time.Minute).Should(Succeed(), "Failed to proxy get")

	_ = s

	return nil
}

var IProxyGet = models.StepDefinition{
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
	Parameters: []models.Parameter{models.URLScheme, models.Reference, models.Port, models.URLPath, models.ProxyGetOptions},
	Function:   IProxyGetFunc,
}

var IProxyGetHTTP = models.StepDefinition{
	Expression: regexp.MustCompile(fmt.Sprintf(`^I proxy get %s%s%s$`, NamedObj, Port, URLPath)),
	Function: func(ctx context.Context, ref, port, path string, options *messages.DataTable) error {
		return IProxyGetFunc(ctx, "", ref, port, path, options)
	},
	Name: "i-proxy-get",
	Text: "I proxy get <reference>:<port><path>",
}
