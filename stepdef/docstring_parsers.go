package stepdef

import (
	"context"
	"errors"
	"reflect"

	messages "github.com/cucumber/messages/go/v21"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func UnmarshalDocString(ctx context.Context, ds *messages.DocString, targetType reflect.Type) (reflect.Value, error) {
	if targetType == reflect.TypeOf((*messages.DocString)(nil)) {
		return reflect.ValueOf(ds), nil
	}
	o := reflect.New(targetType)
	err := yaml.Unmarshal([]byte(ds.Content), o)
	return reflect.ValueOf(o), err
}

// Provides some basic validation object validation
func unmarshalToClientObject(b []byte, targetType reflect.Type) (_ reflect.Value, err error) {
	if !targetType.Implements(reflect.TypeOf((client.Object)(nil))) {
		return reflect.Value{}, errors.New("targetType does not implement client.Object")
	}

	u := &unstructured.Unstructured{}
	err = yaml.Unmarshal(b, u)

	if u.GetAPIVersion() == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty API Version"))
	}
	if u.GetKind() == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty Kind"))
	}
	if u.GetName() == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty Name"))
	}

	if targetType == reflect.TypeOf((*unstructured.Unstructured)(nil)) {
		return reflect.ValueOf(u), nil
	}

	o := reflect.New(targetType)
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, o)
	return reflect.ValueOf(o), err
}

func ParseDocStringToClientObject(ctx context.Context, ds *messages.DocString, targetType reflect.Type) (_ reflect.Value, err error) {
	return unmarshalToClientObject([]byte(ds.Content), targetType)
}
