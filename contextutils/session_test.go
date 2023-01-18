package contextutils

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testernetes/gkube"
)

var _ = Describe("session", func() {

	Context("Saving and Loading", Ordered, func() {
		var ctx context.Context

		BeforeAll(func() {
			ctx = context.Background()
		})

		It("should initialize a session into a ctx", func() {
			ctx = NewPodSessionsFor(ctx)
			Expect(ctx.Value(&session{})).Should(Equal(map[string]*gkube.PodSession{}))
		})

		It("should save into a ctx", func() {
			Save(ctx, "pod", &gkube.PodSession{})
			Expect(ctx.Value(&session{})).Should(Equal(map[string]*gkube.PodSession{
				"pod": &gkube.PodSession{},
			}))
		})

		It("should initialize a session into a ctx", func() {
			pod := Load(ctx, "pod")
			Expect(pod).Should(Equal(&gkube.PodSession{}))
		})
	})

})

func TestSession(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "session Suite")
}
