package arguments

import (
	"encoding/json"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DocString", func() {

	Context("Marshalling", func() {

		It("should run marshal a docstring to a JSON", func() {
			docstring := &DocString{
				DocString: &messages.DocString{
					Content: `{"key1":"val1","key2":12}`,
				},
			}
			json, err := json.Marshal(docstring)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(json).Should(Equal([]byte(`"{\"key1\":\"val1\",\"key2\":12}"`)))
		})

		It("should unmarshal a table into another interface", func() {
			type simple struct {
				Key1 string `json:"key1"`
				Key2 int    `json:"key2"`
			}

			s := &simple{}

			docstring := &DocString{
				DocString: &messages.DocString{
					Content: `{"key1":"val1","key2":12}`,
				},
			}

			Expect(docstring.UnmarshalInto(s)).Should(Succeed())
			Expect(s.Key1).Should(Equal("val1"))
			Expect(s.Key2).Should(Equal(12))
		})
	})

})
