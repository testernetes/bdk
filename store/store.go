package store

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type store struct{}

func Save[T any](ctx context.Context, key string, value T) {
	storage := ctx.Value(&store{}).(map[string]any)

	storage[key] = value
	log.FromContext(ctx).V(1).Info("saved %+v to '%s'\n", value, key)
}

// Load or create new value
func Load[T any](ctx context.Context, key string) T {
	storage := ctx.Value(&store{}).(map[string]any)

	var t T
	log.FromContext(ctx).V(1).Info("loading from '%s' => ", key)
	if value, exists := storage[key]; exists {
		log.FromContext(ctx).V(1).Info("found %+v %T\n", value, value)
		return value.(T)
	}
	log.FromContext(ctx).V(1).Info("did not find\n")

	storage[key] = t
	return t
}

func NewStoreFor(ctx context.Context) context.Context {
	return context.WithValue(ctx, &store{}, make(map[string]any))
}
