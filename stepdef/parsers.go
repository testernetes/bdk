package stepdef

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	CannotParse = "cannot parse '%s' into a %s"
)

func marshalDataTable(dt *messages.DataTable) ([]byte, error) {
	j := map[string]interface{}{}
	for _, row := range dt.Rows {
		if len(row.Cells) != 2 {
			return []byte{}, fmt.Errorf("table must be a width of 2 containing key/value pairs to parse into a struct")
		}
		key := row.Cells[0].Value
		val := row.Cells[1].Value

		// use unmarshal to automatically parse value to correct type
		var v interface{}
		err := yaml.Unmarshal([]byte(val), &v)
		if err != nil {
			return []byte{}, err
		}

		j[key] = v
	}
	return json.Marshal(j)
}

func ParseDataTable(dt *messages.DataTable, targetType reflect.Type) (_ reflect.Value, err error) {
	var o any
	switch targetType.Kind() {
	case reflect.Ptr:
		o = reflect.New(targetType)
		if _, ok := o.(*messages.DataTable); ok {
			o = dt
		} else {
			var b []byte
			b, err = marshalDataTable(dt)
			err = yaml.Unmarshal(b, o)
		}
	default:
		return reflect.Value{}, fmt.Errorf(CannotParse, "DataTable (step argument)", targetType.String())
	}

	return reflect.ValueOf(o), err
}

type stringParsers map[reflect.Type]func(string) (reflect.Value, error)

func (p stringParsers) Parse(s string, t reflect.Type) (reflect.Value, error) {
	parser, supportedType := p[t]
	if !supportedType {

		return reflect.Value{}, fmt.Errorf("No supported parser registered for %s", t)
	}
	return parser(s)
}

var StringParsers = stringParsers{
	reflect.TypeOf(""):          ParseString,
	reflect.TypeOf(int(0)):      ParseInt,
	reflect.TypeOf(int64(0)):    ParseInt64,
	reflect.TypeOf(int32(0)):    ParseInt32,
	reflect.TypeOf(int16(0)):    ParseInt16,
	reflect.TypeOf(int8(0)):     ParseInt8,
	reflect.TypeOf(float64(0)):  ParseFloat64,
	reflect.TypeOf(float32(0)):  ParseFloat32,
	reflect.TypeOf([]byte(nil)): ParseBytes,
	reflect.TypeOf(false):       ParseBool,

	//TODO uint uint8 uint16 uint32 uint64 uintptr

	reflect.TypeOf(time.Duration(0)):                  ParseGeneric(time.ParseDuration),
	reflect.TypeOf((*unstructured.Unstructured)(nil)): ParseUnstructured,
	//reflect.TypeOf(([]client.CreateOption)(nil)):      ParseCreateOptions,
	reflect.TypeOf((types.GomegaMatcher)(nil)): ParseMatcher,
}

func ParseGeneric[T any](f func(string) (T, error)) func(string) (reflect.Value, error) {
	return func(s string) (reflect.Value, error) {
		v, err := f(s)
		return reflect.ValueOf(v), err
	}
}

func ParseString(s string) (reflect.Value, error) {
	return reflect.ValueOf(s), nil
}

func ParseBytes(s string) (reflect.Value, error) {
	return reflect.ValueOf([]byte(s)), nil
}

func ParseBool(s string) (reflect.Value, error) {
	b := strings.Contains(s, "true")
	return reflect.ValueOf(b), nil
}

func ParseUnstructured(s string) (reflect.Value, error) {
	u := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(s), u)

	if u.GetAPIVersion() == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty API Version"))
	}
	if u.GetKind() == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty Kind"))
	}
	if u.GetName() == "" {
		err = errors.Join(errors.New("Provided test case resource has an empty Name"))
	}
	return reflect.ValueOf(u), err
}

func ParseDocString(ds *messages.DocString, targetType reflect.Type) (_ reflect.Value, err error) {
	if targetType == reflect.TypeOf((*messages.DocString)(nil)) {
		return reflect.ValueOf(ds), nil
	}
	return StringParsers.Parse(ds.Content, targetType)
}

func ParseInt(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 0)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int(v)), nil
}

func ParseInt64(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(v), nil
}

func ParseInt32(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int32(v)), nil
}

func ParseInt16(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 16)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int16(v)), nil
}

func ParseInt8(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 8)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int8(v)), nil
}

func ParseFloat64(input string) (reflect.Value, error) {
	v, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(v), nil
}

func ParseFloat32(input string) (reflect.Value, error) {
	v, err := strconv.ParseFloat(input, 32)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(float32(v)), nil
}
