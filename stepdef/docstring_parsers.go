package stepdef

import (
	"context"
	"errors"
	"reflect"

	messages "github.com/cucumber/messages/go/v21"
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
	o, err := toClientObject(targetType)
	if err != nil {
		return reflect.Value{}, err
	}

	err = yaml.Unmarshal(b, o)
	if err != nil {
		return reflect.Value{}, err
	}

	apiVersion, kind := o.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	if apiVersion == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty API Version"))
	}
	if kind == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty Kind"))
	}
	if o.GetName() == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty Name"))
	}
	return reflect.ValueOf(o), err
}

func ParseDocStringToClientObject(ctx context.Context, ds *messages.DocString, targetType reflect.Type) (_ reflect.Value, err error) {
	return unmarshalToClientObject([]byte(ds.Content), targetType)
}
