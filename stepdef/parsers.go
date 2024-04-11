package stepdef

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/store"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	CannotParse = "cannot parse '%s' into a %s"
)

func ParseFileToClientObject(ctx context.Context, path string, targetType reflect.Type) (_ reflect.Value, err error) {
	manifest, err := ioutil.ReadFile(path)
	if err != nil {
		return reflect.Value{}, err
	}
	return unmarshalToClientObject(manifest, targetType)
}

func ParseClientObject(ctx context.Context, s string, targetType reflect.Type) (reflect.Value, error) {
	u := store.Load[*unstructured.Unstructured](ctx, s)
	if targetType == reflect.TypeOf((*unstructured.Unstructured)(nil)) {
		return reflect.ValueOf(u), nil
	}

	o := reflect.New(targetType)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, o)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(o), nil
}

type stringParsers map[reflect.Type]func(context.Context, string) (reflect.Value, error)

func (p stringParsers) Parse(ctx context.Context, s string, t reflect.Type) (reflect.Value, error) {
	if t == nil {
		return reflect.Value{}, fmt.Errorf("target type cannot be nil")
	}

	parser, supportedType := p[t]
	if !supportedType {
		return reflect.Value{}, fmt.Errorf("No supported parser registered for %s", t)
	}
	//fmt.Printf("%s to %T", t.String(), parser)
	return parser(ctx, s)
}

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

	reflect.TypeOf((Assert)(nil)):    parseAssertion,
	reflect.TypeOf(time.Duration(0)): parseGeneric(time.ParseDuration),
	//reflect.TypeOf((*unstructured.Unstructured)(nil)): loadFromStore[*unstructured.Unstructured](),
	//reflect.TypeOf((*corev1.Pod)(nil)):                parsePod,
	reflect.TypeOf((*types.GomegaMatcher)(nil)): Matchers.ParseMatcher,

	reflect.TypeOf(client.DryRunAll):                valueIfTrue(client.DryRunAll),
	reflect.TypeOf(client.FieldOwner("")):           unmarshal[client.FieldOwner],
	reflect.TypeOf(client.GracePeriodSeconds(0)):    unmarshal[client.GracePeriodSeconds],
	reflect.TypeOf(client.PropagationPolicy("")):    unmarshal[client.PropagationPolicy],
	reflect.TypeOf(client.MatchingLabelsSelector{}): parseSelector,
	reflect.TypeOf(client.InNamespace("")):          unmarshal[client.InNamespace],
	reflect.TypeOf(client.Limit(0)):                 unmarshal[client.Limit],
	reflect.TypeOf(client.ForceOwnership):           valueIfTrue(client.ForceOwnership),
	reflect.TypeOf((client.Patch)(nil)):             loadFromStore[client.Patch](),
}

type skipValueErr struct{}

func (e *skipValueErr) Error() string {
	return ""
}

func valueIfTrue[T any](v T) func(ctx context.Context, s string) (reflect.Value, error) {
	return func(ctx context.Context, s string) (reflect.Value, error) {
		if s == "false" {
			return reflect.Value{}, &skipValueErr{}
		}
		return reflect.ValueOf(v), nil
	}
}

func parseDryRun(ctx context.Context, s string) (reflect.Value, error) {
	return reflect.ValueOf(client.DryRunAll), nil
}

func parseSelector(ctx context.Context, s string) (reflect.Value, error) {
	selector, err := labels.Parse(s)
	return reflect.ValueOf(selector.(client.MatchingLabelsSelector)), err
}

func unmarshal[T any](ctx context.Context, s string) (reflect.Value, error) {
	var t *T
	err := yaml.Unmarshal([]byte(s), t)
	return reflect.ValueOf(t), err
}

func parseGeneric[T any, S string | []byte](f func(S) (T, error)) func(context.Context, string) (reflect.Value, error) {
	var s S
	switch any(s).(type) {
	case string:
		return func(ctx context.Context, s string) (reflect.Value, error) {
			v, err := f(any(s).(S))
			return reflect.ValueOf(v), err
		}
	case []byte:
		return func(ctx context.Context, s string) (reflect.Value, error) {
			b := []byte(s)
			v, err := f(any(b).(S))
			return reflect.ValueOf(v), err
		}
	}
	panic("cannot get here")
}

func parseAny(ctx context.Context, s string) (reflect.Value, error) {
	var v any
	err := yaml.Unmarshal([]byte(s), v)
	return reflect.ValueOf(v), err
}

func parseString(ctx context.Context, s string) (reflect.Value, error) {
	return reflect.ValueOf(s), nil
}

func parseBytes(ctx context.Context, s string) (reflect.Value, error) {
	return reflect.ValueOf([]byte(s)), nil
}

func parseBool(ctx context.Context, s string) (reflect.Value, error) {
	b := strings.Contains(s, "true")
	return reflect.ValueOf(b), nil
}

func parsePod(ctx context.Context, s string) (reflect.Value, error) {
	u := store.Load[*unstructured.Unstructured](ctx, s)
	pod := &corev1.Pod{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructuredWithValidation(u.Object, pod, false)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(pod), nil
}

func parseInt(ctx context.Context, input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 0)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int(v)), nil
}

func parseInt64(ctx context.Context, input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(v), nil
}

func parseInt32(ctx context.Context, input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int32(v)), nil
}

func parseInt16(ctx context.Context, input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 16)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int16(v)), nil
}

func parseInt8(ctx context.Context, input string) (reflect.Value, error) {
	v, err := strconv.ParseInt(input, 10, 8)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(int8(v)), nil
}

func parseFloat64(ctx context.Context, input string) (reflect.Value, error) {
	v, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(v), nil
}

func parseFloat32(ctx context.Context, input string) (reflect.Value, error) {
	v, err := strconv.ParseFloat(input, 32)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(float32(v)), nil
}

func parseAssertion(ctx context.Context, phrase string) (reflect.Value, error) {
	if contains(phrase, EventuallyPhrases) {
		return reflect.ValueOf(Eventually), nil
	}
	if contains(phrase, ConsistentlyPhrases) {
		return reflect.ValueOf(Consistently), nil
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

func loadFromStore[T any]() func(context.Context, string) (reflect.Value, error) {
	return func(ctx context.Context, s string) (reflect.Value, error) {
		return reflect.ValueOf(store.Load[T](ctx, s)), nil
	}
}
