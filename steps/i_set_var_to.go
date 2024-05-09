package steps

import (
	"context"
	"fmt"
	"reflect"

	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ISetVar = stepdef.StepDefinition{
	Name:     "i-set-var",
	Text:     "^I set {var} to {text}$",
	Function: SetVarFunc,
	StepArg:  stepdef.NoStepArg,
	Help:     `Assigns a value to a variable`,
}

var ISetVarFromJSONPath = stepdef.StepDefinition{
	Name:     "i-set-var",
	Text:     "^I set {var} from {reference} jsonpath {jsonpath}$",
	Function: SetVarFromJSONPathFunc,
	StepArg:  stepdef.NoStepArg,
	Help:     `Assigns a value to a variable`,
}

func FromJSONPathFunc(o client.Object, jp string) ([][]reflect.Value, error) {
	j := jsonpath.New("")
	if err := j.Parse(jp); err != nil {
		return nil, fmt.Errorf("JSON Path '%s' is invalid: %s", jp, err.Error())
	}

	var obj interface{}
	if u, ok := o.(*unstructured.Unstructured); ok {
		obj = u.UnstructuredContent()
	} else {
		obj = o
	}

	results, err := j.FindResults(obj)
	if err != nil {
		return nil, fmt.Errorf("JSON Path '%s' failed: %s", jp, err.Error())
	}

	return results, nil
}

var SetVarFromJSONPathFunc = func(ctx context.Context, key string, ref *unstructured.Unstructured, jsonpath string) (err error) {
	values, err := FromJSONPathFunc(ref, jsonpath)
	if !(len(values) == 1 && len(values[0]) == 1) {
		return fmt.Errorf("jsonpath should resolve to a single variable")
	}

	key = "scn-var-" + key
	store.Save(ctx, key, fmt.Sprint(values[0][0].Interface()))
	return nil
}

var SetVarFunc = func(ctx context.Context, key, value string) (err error) {
	key = "scn-var-" + key
	store.Save(ctx, key, value)
	return nil
}
