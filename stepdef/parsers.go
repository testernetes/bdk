package stepdef

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
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

func ParseDataTable(ctx context.Context, dt *messages.DataTable, targetType reflect.Type) (_ reflect.Value, err error) {
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

func (p stringParsers) Parse(ctx context.Context, s string, t reflect.Type) (reflect.Value, error) {
	parser, supportedType := p[t]
	if !supportedType {

		return reflect.Value{}, fmt.Errorf("No supported parser registered for %s", t)
	}
	return parser(s)
}

var i interface{}

var StringParsers = stringParsers{
	reflect.TypeOf(""):          parseString,
	reflect.TypeOf(int(0)):      parseInt,
	reflect.TypeOf(int64(0)):    parseInt64,
	reflect.TypeOf(int32(0)):    parseInt32,
	reflect.TypeOf(int16(0)):    parseInt16,
	reflect.TypeOf(int8(0)):     parseInt8,
	reflect.TypeOf(float64(0)):  parseFloat64,
	reflect.TypeOf(float32(0)):  parseFloat32,
	reflect.TypeOf([]byte(nil)): parseBytes,
	reflect.TypeOf(false):       parseBool,

	//TODO uint uint8 uint16 uint32 uint64 uintptr

	reflect.TypeOf(Assertion("")):                     parseAssertion,
	reflect.TypeOf(time.Duration(0)):                  parseGeneric(time.ParseDuration),
	reflect.TypeOf((*unstructured.Unstructured)(nil)): parseUnstructured,
	reflect.TypeOf((*corev1.Pod)(nil)):                parseUnstructured,
	reflect.TypeOf((*corev1.PodSession)(nil)):         parsePodSession,
	reflect.TypeOf((types.GomegaMatcher)(nil)):        Matchers.ParseMatcher,
	//reflect.TypeOf(([]client.CreateOption)(nil)):      ParseCreateOptions,

	// best effort json unmarshalling to interface
	reflect.TypeOf(i): parseAny,
}

func parseGeneric[T any, S string | []byte](f func(S) (T, error)) func(string) (reflect.Value, error) {
	var s S
	switch any(s).(type) {
	case string:
		return func(s string) (reflect.Value, error) {
			v, err := f(any(s).(S))
			return reflect.ValueOf(v), err
		}
	case []byte:
		return func(s string) (reflect.Value, error) {
			b := []byte(s)
			v, err := f(any(b).(S))
			return reflect.ValueOf(v), err
		}
	}
	panic("cannot get here")
}

func parseAny(s string) (reflect.Value, error) {
	var v any
	err := yaml.Unmarshal([]byte(s), v)
	return reflect.ValueOf(v), err
}

func parseString(s string) (reflect.Value, error) {
	return reflect.ValueOf(s), nil
}

func parseBytes(s string) (reflect.Value, error) {
	return reflect.ValueOf([]byte(s)), nil
}

func parseBool(s string) (reflect.Value, error) {
	b := strings.Contains(s, "true")
	return reflect.ValueOf(b), nil
}

func parseUnstructured(s string) (reflect.Value, error) {
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

func parseDocString(ctx context.Context, ds *messages.DocString, targetType reflect.Type) (_ reflect.Value, err error) {
	if targetType == reflect.TypeOf((*messages.DocString)(nil)) {
		return reflect.ValueOf(ds), nil
	}
	if targetType == reflect.TypeOf((*unstructured.Unstructured)(nil)) {
		return parseUnstructured(ds.Content)
	}
	return StringParsers.Parse(ctx, ds.Content, targetType)
}

func parseInt(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 0)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int(v)), nil
}

func parseInt64(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(v), nil
}

func parseInt32(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int32(v)), nil
}

func parseInt16(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 16)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int16(v)), nil
}

func parseInt8(input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 8)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int8(v)), nil
}

func parseFloat64(input string) (reflect.Value, error) {
	v, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(v), nil
}

func parseFloat32(input string) (reflect.Value, error) {
	v, err := strconv.ParseFloat(input, 32)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(float32(v)), nil
}

func parseAssertion(phrase string) (reflect.Value, error) {
	if contains(phrase, EventuallyPhrases) {
		return reflect.ValueOf(EventuallyAssertion), nil
	}
	if contains(phrase, ConsistentlyPhrases) {
		return reflect.ValueOf(ConsistentlyAssertion), nil
	}
	return reflect.Value{}, errors.New("cannot determine if eventually or consistently")
}

func contains(s string, a []string) bool {
	for i := range a {
		if a[i] == s {
			return true
		}
	}
	return false
}
