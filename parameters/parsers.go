package parameters

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"

	"github.com/testernetes/bdk/arguments"
)

const (
	CannotParse = "cannot parse '%s' into a %s"
)

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

func DocStringParseString(docString *arguments.DocString, targetType reflect.Type) (reflect.Value, error) {
	if docString == nil {
		return reflect.ValueOf(""), nil
	}
	return reflect.ValueOf(docString.Content), nil
}

func ParseWriter(input string, targetType reflect.Type) (reflect.Value, error) {
	var writer io.Writer

	writer = os.Stdout

	return reflect.ValueOf(writer), nil
}
