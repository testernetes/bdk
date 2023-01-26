package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tag filtering", func() {

	When("Filtering for a specific tag", func() {
		filter := Tag{"tag", true}

		It("should pass when it finds that tag", Focus, func() {
			Expect(Tag{"tag", true}.passes(filter)).Should(BeTrue())
		})
		It("should fail when it does not find that tag", Focus, func() {
			Expect(Tag{"other", true}.passes(filter)).Should(BeFalse())
		})
	})

	When("Filtering out a specific tag", func() {
		filter := Tag{"tag", false}

		It("should fail when it finds that tag", Focus, func() {
			Expect(Tag{"tag", true}.passes(filter)).Should(BeFalse())
		})
		It("should pass when it does not find that tag", Focus, func() {
			Expect(Tag{"other", true}.passes(filter)).Should(BeTrue())
		})
	})

})
