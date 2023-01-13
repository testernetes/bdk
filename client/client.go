package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/testernetes/gkube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type helper struct{}

func NewClientFor(ctx context.Context, opts ...gkube.HelperOption) context.Context {
	client := gkube.NewKubernetesHelper(opts...)
	rctx := context.WithValue(ctx, helper{}, client)
	return rctx
}

func MustGetClientFrom(ctx context.Context) gkube.KubernetesHelper {
	v := ctx.Value(helper{})
	if v == nil {
		panic(errors.New("no client initialized"))
	}
	if client, ok := v.(gkube.KubernetesHelper); ok {
		return client
	}
	panic(errors.New(fmt.Sprintf("expected a client found %T in ctx", v)))
}

func PodLogOptionsFrom(table *messages.DataTable) *corev1.PodLogOptions {
	opts := &corev1.PodLogOptions{}
	if table == nil {
		return opts
	}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		opt := row.Cells[0].Value
		val := row.Cells[1].Value
		switch opt {
		case "container":
			opts.Container = val
		case "follow":
			opts.Follow = val == "true"
		case "previous":
			opts.Previous = val == "true"
		case "since seconds":
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				panic(err)
			}
			opts.SinceSeconds = &v
		case "since time":
			panic("since time not yet implemented")
		case "timestamps":
			opts.Timestamps = val == "true"
		case "tail lines":
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				panic(err)
			}
			opts.TailLines = &v
		case "limit bytes":
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				panic(err)
			}
			opts.LimitBytes = &v
		}
	}
	return opts
}

func ProxyGetOptionsFrom(table *messages.DataTable) map[string]string {
	params := make(map[string]string)
	if table == nil {
		return params
	}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		opt := row.Cells[0].Value
		val := row.Cells[1].Value
		params[opt] = val
	}
	return params
}

func PatchOptionsFrom(o client.Object, table *messages.DataTable) []interface{} {
	opts := []interface{}{o}
	if table == nil {
		return opts
	}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		opt := row.Cells[0].Value
		val := row.Cells[1].Value
		switch opt {
		case "dry run":
			if val == "true" {
				opts = append(opts, client.DryRunAll)
			}
		case "field owner":
			opts = append(opts, client.FieldOwner(val))
		case "force":
			if val == "true" {
				opts = append(opts, client.ForceOwnership)
			}
		case "patch":
			patch, err := yaml.YAMLToJSON([]byte(val))
			if err != nil {
				panic(err)
			}
			opts = append(opts, client.RawPatch(types.StrategicMergePatchType, patch))
		}
	}
	return opts
}

func CreateOptionsFrom(o client.Object, table *messages.DataTable) []interface{} {
	opts := []interface{}{o}
	if table == nil {
		return opts
	}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		opt := row.Cells[0].Value
		val := row.Cells[1].Value
		switch opt {
		case "dry run":
			if val == "true" {
				opts = append(opts, client.DryRunAll)
			}
		case "field owner":
			opts = append(opts, client.FieldOwner(val))
		default:
			panic(fmt.Sprintf("invalid create option %s", opt))
		}
	}
	return opts
}

func DeleteOptionsFrom(o client.Object, table *messages.DataTable) []interface{} {
	opts := []interface{}{o}
	if table == nil {
		return opts
	}
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		opt := row.Cells[0].Value
		val := row.Cells[1].Value
		switch opt {
		case "dry run":
			if val == "true" {
				opts = append(opts, client.DryRunAll)
			}
		case "grace period seconds":
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				panic(err)
			}
			opts = append(opts, client.GracePeriodSeconds(v))
		case "preconditions":
			panic("preconditions not yet supported")
		case "propagation policy":
			opts = append(opts, client.PropagationPolicy(val))
		default:
			panic(fmt.Sprintf("invalid delete option %s", opt))
		}
	}
	return opts
}
