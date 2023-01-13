package arguments

import (
	"errors"

	messages "github.com/cucumber/messages/go/v21"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

var (
	ErrNoAPIVersion = errors.New("Provided test case resource has an empty API Version")
	ErrNoKind       = errors.New("Provided test case resource has an empty Kind")
	ErrNoName       = errors.New("Provided test case resource has an empty Name")
)

type DocString struct {
	*messages.DocString
}

func (d *DocString) UnmarshalInto(o interface{}) error {
	return yaml.Unmarshal([]byte(d.Content), o)
}

func (d *DocString) GetUnstructured() (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(d.Content), u)
	if err != nil {
		return u, err
	}

	if u.GetAPIVersion() == "" {
		return u, ErrNoAPIVersion
	}
	if u.GetKind() == "" {
		return u, ErrNoKind
	}
	if u.GetName() == "" {
		return u, ErrNoName
	}
	return u, nil
}
