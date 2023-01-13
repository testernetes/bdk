package sessions

import (
	"context"

	"github.com/testernetes/gkube"
)

type sessions struct{}

func NewPodSessionsFor(ctx context.Context) context.Context {
	s := make(map[string]*gkube.PodSession)
	rctx := context.WithValue(ctx, &sessions{}, s)
	return rctx
}

func Save(ctx context.Context, key string, value *gkube.PodSession) {
	v := ctx.Value(&sessions{})
	if session, ok := v.(map[string]*gkube.PodSession); ok {
		session[key] = value
	}
}

func Load(ctx context.Context, key string) *gkube.PodSession {
	v := ctx.Value(&sessions{})
	if session, ok := v.(map[string]*gkube.PodSession); ok {
		return session[key]
	}
	return nil
}
