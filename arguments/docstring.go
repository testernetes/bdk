package arguments

import (
	"bytes"
	"errors"

	messages "github.com/cucumber/messages/go/v21"
	"k8s.io/apimachinery/pkg/util/json"
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
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(d.Content)
	return buf.Bytes(), err
}

func (d *DocString) UnmarshalJSON(b []byte) error {
	if b != nil && d != nil {
		d.DocString = &messages.DocString{
			Content: string(b),
		}
	}
	return nil
}

func (d *DocString) UnmarshalInto(o interface{}) error {
	return yaml.Unmarshal([]byte(d.Content), o)
}
