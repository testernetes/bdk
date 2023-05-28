package contextutils

import (
	"context"
	"errors"
	"fmt"

	"github.com/testernetes/gkube"
)

type helper struct{}

var hotClient gkube.KubernetesHelper

func NewClientFor(ctx context.Context, opts ...gkube.HelperOption) context.Context {
	if hotClient == nil {
		hotClient = gkube.NewKubernetesHelper(opts...)
	}
	rctx := context.WithValue(ctx, helper{}, hotClient)
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
