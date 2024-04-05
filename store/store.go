package store

import (
	"context"
	"fmt"
)

type store struct{}

func Save[T any](ctx context.Context, key string, value T) {
	storage := ctx.Value(&store{}).(map[string]any)

	key = fmt.Sprintf("%T %s", value, key)
	storage[key] = value
}

// Load or create new value
func Load[T any](ctx context.Context, key string) T {
	storage := ctx.Value(&store{}).(map[string]any)

	var t T
	key = fmt.Sprintf("%T %s", t, key)
	if value, exists := storage[key]; exists {
		return value.(T)
	}

	storage[key] = t
	return t
}

func NewStoreFor(ctx context.Context) context.Context {
	return context.WithValue(ctx, &store{}, make(map[string]any))
}
