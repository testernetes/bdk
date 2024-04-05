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
	"github.com/testernetes/bdk/store"
	"github.com/testernetes/gkube"
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

type stringParsers map[reflect.Type]func(context.Context, string) (reflect.Value, error)

func (p stringParsers) Parse(ctx context.Context, s string, t reflect.Type) (reflect.Value, error) {
	parser, supportedType := p[t]
	if !supportedType {
		return reflect.Value{}, fmt.Errorf("No supported parser registered for %s", t)
	}
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

	reflect.TypeOf((Assert)(nil)):                     parseAssertion,
	reflect.TypeOf(time.Duration(0)):                  parseGeneric(time.ParseDuration),
	reflect.TypeOf((*unstructured.Unstructured)(nil)): loadFromStore[*unstructured.Unstructured](),
	reflect.TypeOf((*corev1.Pod)(nil)):                parsePod,
	reflect.TypeOf(gkube.PodSession{}):                loadFromStore[gkube.PodSession](),
	reflect.TypeOf((types.GomegaMatcher)(nil)):        Matchers.ParseMatcher,

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

var clientOptions = map[string]reflect.Type{
	"FieldOwner":         reflect.TypeOf(client.FieldOwner("")),
	"DryRun":             reflect.TypeOf(client.DryRunAll),
	"GracePeriodSeconds": reflect.TypeOf(client.GracePeriodSeconds(0)),
	"PropagationPolicy":  reflect.TypeOf(client.PropagationPolicy("")),
	"Selector":           reflect.TypeOf(client.MatchingLabelsSelector{}),
	"Namespace":          reflect.TypeOf(client.InNamespace("")),
	"Limit":              reflect.TypeOf(client.Limit(0)),
	"Force":              reflect.TypeOf(client.ForceOwnership),
	// TODO FieldSelector
}

func parseClientOptions(ctx context.Context, dt *messages.DataTable, targetType reflect.Type) (reflect.Value, error) {
	if targetType.Kind() != reflect.Slice {
		return reflect.Value{}, errors.New("expected targetType to be a slice of client options")
	}

	opts := reflect.New(targetType)

	// switch from []client.CreateOption to client.CreateOption for example
	targetType = targetType.Elem()

	for _, r := range dt.Rows {
		if len(r.Cells) != 2 {
			return reflect.Value{}, errors.New("expected table to have width of 2")
		}

		key := r.Cells[0].Value
		clientOption := clientOptions[key]
		if !clientOption.Implements(targetType) {
			return reflect.Value{}, fmt.Errorf("%s is not a valid option for %s", key, targetType.String())
		}

		value := r.Cells[1].Value
		opt, err := StringParsers[clientOption](ctx, value)
		if err != nil {
			if _, skip := err.(*skipValueErr); skip {
				continue
			}
			return reflect.Value{}, err
		}
		opts = reflect.Append(opts, opt)
	}
	return opts, nil
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
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, pod)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(pod), nil
}

func loadFromStore[T any]() func(context.Context, string) (reflect.Value, error) {
	return func(ctx context.Context, s string) (reflect.Value, error) {
		u := store.Load[T](ctx, s)
		return reflect.ValueOf(u), nil
	}
}

func parseDocString(ctx context.Context, ds *messages.DocString, targetType reflect.Type) (_ reflect.Value, err error) {
	if targetType == reflect.TypeOf((*messages.DocString)(nil)) {
		return reflect.ValueOf(ds), nil
	}
	if targetType == reflect.TypeOf((*unstructured.Unstructured)(nil)) {
		u := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(ds.Content), u)

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
	return StringParsers.Parse(ctx, ds.Content, targetType)
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
