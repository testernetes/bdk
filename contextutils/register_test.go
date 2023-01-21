package contextutils

import (
	"context"

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
			SaveObject(ctx, "obj", &unstructured.Unstructured{})
			Expect(ctx.Value(&register{})).Should(Equal(map[string]*unstructured.Unstructured{
				"obj": &unstructured.Unstructured{},
			}))
		})

		It("should initialize a register into a ctx", func() {
			obj := LoadObject(ctx, "obj")
			Expect(obj).Should(Equal(&unstructured.Unstructured{}))
		})
	})

})
