package store

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type store struct{}

func Save[T any](ctx context.Context, key string, value T) {
	storage := ctx.Value(&store{}).(map[string]any)

	storage[key] = value
	log.FromContext(ctx).V(1).Info("Store Saved", "Key", key, "Value", value)
}

// Load or create new value
func Load[T any](ctx context.Context, key string) T {
	storage := ctx.Value(&store{}).(map[string]any)

	var t T
	if value, exists := storage[key]; exists {
		log.FromContext(ctx).V(1).Info("Store Loaded", "Key", key, "Value", value)
		return value.(T)
	}
	log.FromContext(ctx).V(1).Info("Store Loaded", "Key", key, "Value", "NotFound")

	storage[key] = t
	return t
}

func NewStoreFor(ctx context.Context) context.Context {
	return context.WithValue(ctx, &store{}, make(map[string]any))
}
