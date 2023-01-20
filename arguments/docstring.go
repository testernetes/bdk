package arguments

import (
	"errors"

	messages "github.com/cucumber/messages/go/v21"
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

func (d *DocString) MarshalJSON() ([]byte, error) {
	return []byte(d.Content), nil
}

func (d *DocString) UnmarshalJSON(b []byte) error {
	d.Content = string(b)
	return nil
}

func (d *DocString) UnmarshalInto(o interface{}) error {
	return yaml.Unmarshal([]byte(d.Content), o)
}
