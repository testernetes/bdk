package contextutils

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type register struct{}

func NewRegisterFor(ctx context.Context) context.Context {
	objRegister := make(map[string]*unstructured.Unstructured)
	rctx := context.WithValue(ctx, &register{}, objRegister)
	return rctx
}

func SaveObject(ctx context.Context, key string, value *unstructured.Unstructured) {
	v := ctx.Value(&register{})
	if objRegister, ok := v.(map[string]*unstructured.Unstructured); ok {
		objRegister[key] = value
	}
}

func LoadObject(ctx context.Context, key string) *unstructured.Unstructured {
	v := ctx.Value(&register{})
	objRegister, ok := v.(map[string]*unstructured.Unstructured)
	if !ok {
		panic("no object store")
	}

	if u, ok := objRegister[key]; ok {
		return u
	}

	objRegister[key] = &unstructured.Unstructured{}
	return objRegister[key]
}

func LoadPod(ctx context.Context, key string) *corev1.Pod {
	u := LoadObject(ctx, key)
	if u == nil {
		return nil
	}
	pod := &corev1.Pod{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), pod)
	if err != nil {
		panic(fmt.Errorf("Could not load %s as a Pod: %w", key, err))
	}
	return pod
}
