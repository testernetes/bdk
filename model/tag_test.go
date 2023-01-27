package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tag filtering", func() {

	When("Filtering for a specific tag", func() {
		f := Filter{"tag", true}

		It("should pass when it finds that tag", func() {
			Expect(f.filters(Tag{"tag"})).Should(BeTrue())
		})
		It("should fail when it does not find that tag", func() {
			Expect(f.filters(Tag{"other"})).Should(BeFalse())
		})
	})

	When("Filtering out a specific tag", func() {
		f := Filter{"tag", false}

		It("should fail when it finds that tag", func() {
			Expect(f.filters(Tag{"tag"})).Should(BeFalse())
		})
		It("should pass when it does not find that tag", func() {
			Expect(f.filters(Tag{"other"})).Should(BeTrue())
		})
	})

})
