package stepdef

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	messages "github.com/cucumber/messages/go/v21"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func marshalDataTable(dt *messages.DataTable) ([]byte, error) {
	j := map[string]interface{}{}
	for _, row := range dt.Rows {
		if len(row.Cells) != 2 {
			return []byte{}, fmt.Errorf("table must be a width of 2 containing key/value pairs to parse into a struct")
		}
		key := row.Cells[0].Value
		val := row.Cells[1].Value

		// use unmarshal for best effort parse of json value to correct type
		var v interface{}
		err := yaml.Unmarshal([]byte(val), &v)
		if err != nil {
			return []byte{}, err
		}

		j[key] = v
	}
	return json.Marshal(j)
}

func UnmarshalDataTable(ctx context.Context, dt *messages.DataTable, targetType reflect.Type) (_ reflect.Value, err error) {
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

func ParseClientOptions(ctx context.Context, dt *messages.DataTable, targetType reflect.Type) (reflect.Value, error) {
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
