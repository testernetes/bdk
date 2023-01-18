package contextutils

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Register", func() {

	Context("Saving and Loading", Ordered, func() {
		var ctx context.Context

		BeforeAll(func() {
			ctx = context.Background()
		})

		It("should initialize a register into a ctx", func() {
			ctx = NewRegisterFor(ctx)
			Expect(ctx.Value(&register{})).Should(Equal(map[string]*unstructured.Unstructured{}))
		})

		It("should save into a ctx", func() {
			Save(ctx, "pod", &unstructured.Unstructured{})
			Expect(ctx.Value(&register{})).Should(Equal(map[string]*unstructured.Unstructured{
				"pod": &unstructured.Unstructured{},
			}))
		})

		It("should initialize a register into a ctx", func() {
			pod := Load(ctx, "pod")
			Expect(pod).Should(Equal(&unstructured.Unstructured{}))
		})
	})

})

func TestRegister(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Register Suite")
}
