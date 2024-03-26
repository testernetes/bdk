package contextutils

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type variables struct{}

func NewVariablesFor(ctx context.Context) context.Context {
	objRegister := make(map[string]*unstructured.Unstructured)
	rctx := context.WithValue(ctx, &register{}, objRegister)
	return rctx
}

func SaveVariable(ctx context.Context, key string, value string) {
	v := ctx.Value(&variables{})
	if vars, ok := v.(map[string]string); ok {
		vars[key] = value
	}
}

func LoadVariable(ctx context.Context, key string) string {
	v := ctx.Value(&variables{})
	if vars, ok := v.(map[string]string); ok {
		return vars[key]
	}
	return ""
}
