package stepdef

//import (
//	messages "github.com/cucumber/messages/go/v21"
//	. "github.com/onsi/ginkgo/v2"
//	. "github.com/onsi/gomega"
//)
//
//var _ = Describe("Datatable", func() {
//
//	Context("JSON Marshalling and Unmarshalling", func() {
//
//		It("should run marshal a table to a JSON", func() {
//			table := &messages.DataTable{
//				Rows: []*messages.TableRow{
//					{
//						Cells: []*messages.TableCell{{Value: "key1"}, {Value: "val1"}},
//					},
//					{
//						Cells: []*messages.TableCell{{Value: "key2"}, {Value: "12"}},
//					},
//				},
//			}
//			json, err := marshalDataTable(table)
//			Expect(err).ShouldNot(HaveOccurred())
//			Expect(json).Should(Equal([]byte(`{"key1":"val1","key2":12}`)))
//		})
//	})
//})
