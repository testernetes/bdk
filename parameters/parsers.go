package parameters

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"

	messages "github.com/cucumber/messages/go/v21"
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
			return []byte{}, ErrTableMustBeWidthTwo
		}
		key := row.Cells[0].Value
		val := row.Cells[1].Value

		// Get the right type
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

func ParseDocString(ds *messages.DocString, targetType reflect.Type) (_ reflect.Value, err error) {
	var o any
	switch targetType.Kind() {
	case reflect.Ptr:
		o = reflect.New(targetType)
		switch u := o.(type) {
		case *messages.DocString:
			o = ds
		case *unstructured.Unstructured:
			if u.GetAPIVersion() == "" {
				return reflect.ValueOf(u), errors.New("Provided test case resource has an empty API Version")
			}
			if u.GetKind() == "" {
				return reflect.ValueOf(u), errors.New("Provided test case resource has an empty Kind")
			}
			if u.GetName() == "" {
				return reflect.ValueOf(u), errors.New("Provided test case resource has an empty Name")
			}
		default:
			b := []byte(ds.Content)
			err = yaml.Unmarshal(b, o)
		}
	case reflect.String:
		o = ds.Content
	default:
		return reflect.Value{}, fmt.Errorf(CannotParse, "DocString (step argument)", targetType.String())
	}

	return reflect.ValueOf(o), err
}

func ParseNumber(input string, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.Int:
		return ParseInt(input)
	case reflect.Int64:
		return ParseInt64(input)
	case reflect.Int32:
		return ParseInt32(input)
	case reflect.Int16:
		return ParseInt16(input)
	case reflect.Int8:
		return ParseInt8(input)
	case reflect.Float64:
		return ParseFloat64(input)
	case reflect.Float32:
		return ParseFloat32(input)
	}
	return reflect.Value{}, fmt.Errorf(CannotParse, input, targetType.String())
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

func ParseString(input string, targetType reflect.Type) (reflect.Value, error) {
	if targetType.Kind() == reflect.String {
		return reflect.ValueOf(input), nil
	}
	if targetType.Kind() == reflect.Slice {
		if targetType == reflect.TypeOf([]byte(nil)) {
			return reflect.ValueOf([]byte(input)), nil
		}
	}
	return reflect.Value{}, fmt.Errorf(CannotParse, input, targetType.String())
}

func ParseArray(input string, targetType reflect.Type) (reflect.Value, error) {
	if targetType.Kind() != reflect.Slice {
		return reflect.Value{}, errors.New("Must be slice")
	}
	switch targetType {
	case reflect.TypeOf([]byte(nil)):
		return reflect.ValueOf([]byte(input)), nil
	}
	return reflect.Value{}, errors.New("Must be slice")
}

func ParseWriter(input string, targetType reflect.Type) (reflect.Value, error) {
	var writer io.Writer

	writer = os.Stdout

	return reflect.ValueOf(writer), nil
}
