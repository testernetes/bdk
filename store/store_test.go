package store

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Store", func() {

	Context("Saving and Loading", Ordered, func() {
		var ctx context.Context

		BeforeAll(func() {
			ctx = context.Background()
		})

		It("should initialize a store into a ctx", func() {
			ctx = NewStoreFor(ctx)
			Expect(ctx.Value(&store{})).Should(Equal(map[string]any{}))
		})

		It("should save into a ctx", func() {
			Save(ctx, "obj", &unstructured.Unstructured{})
			Expect(ctx.Value(&store{})).Should(Equal(map[string]any{
				"*unstructured.Unstructured obj": &unstructured.Unstructured{},
			}))
		})

		It("should load from a ctx", func() {
			var u *unstructured.Unstructured
			u = Load[*unstructured.Unstructured](ctx, "obj")
			Expect(u).Should(Equal(&unstructured.Unstructured{}))
		})
	})

})
