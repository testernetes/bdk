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
		var u *unstructured.Unstructured

		BeforeAll(func() {
			ctx = context.Background()
			u = &unstructured.Unstructured{}
			u.SetName("me")
		})

		It("should initialize a store into a ctx", func() {
			ctx = NewStoreFor(ctx)
			Expect(ctx.Value(&store{})).Should(Equal(map[string]any{}))
		})

		It("should save into a ctx", func() {
			Save(ctx, "obj", u)
			Expect(ctx.Value(&store{})).Should(Equal(map[string]any{
				"obj": u,
			}))
		})

		It("should load from a ctx", func() {
			u := Load[*unstructured.Unstructured](ctx, "obj")
			Expect(u.GetName()).Should(Equal("me"))
			u.SetName("bar")
		})

		It("should load from a ctx", func() {
			u := Load[*unstructured.Unstructured](ctx, "obj")
			Expect(u.GetName()).Should(Equal("bar"))
		})
	})

})
