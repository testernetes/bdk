package arguments

import (
	"encoding/json"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Datatable", func() {

	Context("JSON Marshalling and Unmarshalling", func() {

		It("should run marshal a table to a JSON", func() {
			table := &DataTable{
				DataTable: &messages.DataTable{
					Rows: []*messages.TableRow{
						{
							Cells: []*messages.TableCell{{Value: "key1"}, {Value: "val1"}},
						},
						{
							Cells: []*messages.TableCell{{Value: "key2"}, {Value: "12"}},
						},
					},
				},
			}
			json, err := json.Marshal(table)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(json).Should(Equal([]byte(`{"key1":"val1","key2":12}`)))
		})

		It("should unmarshal a table into another interface", func() {
			table := &DataTable{
				DataTable: &messages.DataTable{
					Rows: []*messages.TableRow{
						{
							Cells: []*messages.TableCell{{Value: "key1"}, {Value: "val1"}},
						},
						{
							Cells: []*messages.TableCell{{Value: "key2"}, {Value: "12"}},
						},
					},
				},
			}
			type simple struct {
				Key1 string `json:"key1"`
				Key2 int    `json:"key2"`
			}

			s := &simple{}

			Expect(table.UnmarshalInto(s)).Should(Succeed())
			Expect(s.Key1).Should(Equal("val1"))
			Expect(s.Key2).Should(Equal(12))
		})
	})

})
