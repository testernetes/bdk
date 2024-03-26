package steps

import (
	"context"

	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/model"
	"github.com/testernetes/bdk/parameters"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	sd, err := model.NewStepDefinition(
		"a-resource-from-file",
		"a <reference> from <filename>",
		`Assigns a reference to the resource given in the filename. This reference can be referred to
in future steps in the same scenario. JSON and YAML formats are accepted.`,
		`
	Given cm from config.yaml:`,
		SaveObjectFunc,
		parameters.NoStepArg,
	)
	if err != nil {
		panic(err)
	}
	model.Default.Add(sd)
	AResource, err := model.NewStepDefinition(
		"a-resource",
		"a resource called <reference>",
		`Assigns a reference to the resource given in the DocString. This reference can be referred to
in future steps in the same scenario. JSON and YAML formats are accepted.`,
		`Given a resource called cm:
	  """
	  apiVersion: v1
	  kind: ConfigMap
	  metadata:
	    name: example
	    namespace: default
	  data:
	    foo: bar
	  """`,
		SaveObjectFunc,
		parameters.Manifest,
	)
	if err != nil {
		panic(err)
	}
	model.Default.Add(AResource)
}

var SaveObjectFunc = func(ctx context.Context, ref string, u *unstructured.Unstructured) (err error) {
	contextutils.SaveObject(ctx, ref, u)
	return nil
}
