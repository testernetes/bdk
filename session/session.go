package session

import (
	"context"

	"github.com/testernetes/gkube"
)

type session struct{}

func NewPodSessionsFor(ctx context.Context) context.Context {
	s := make(map[string]*gkube.PodSession)
	rctx := context.WithValue(ctx, &session{}, s)
	return rctx
}

func Save(ctx context.Context, key string, value *gkube.PodSession) {
	v := ctx.Value(&session{})
	if session, ok := v.(map[string]*gkube.PodSession); ok {
		session[key] = value
	}
}

func Load(ctx context.Context, key string) *gkube.PodSession {
	v := ctx.Value(&session{})
	if session, ok := v.(map[string]*gkube.PodSession); ok {
		return session[key]
	}
	return nil
}
