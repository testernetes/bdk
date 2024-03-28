package stepdef

import (
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	DescribeTable("Parsing a StringParameter match",
		func(p StringParameter, match string, t reflect.Type, expected interface{}) {
			val, err := p.Parser(match, t)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(val.Interface()).Should(Equal(expected))
		},
		Entry("Text Parser", Text, "word", reflect.TypeOf(""), "word"),
		Entry("Number Parser", Number, "47", reflect.TypeOf(1), 47),
		//Entry("Number Parser", Number, "47", int(47)),
		//Entry("When arg is a int", "a 47", 47, func(ctx context.Context, a int) error { return nil }),
		//Entry("When arg is a int8", "a 8", int8(8), func(ctx context.Context, a int8) error { return nil }),
		//Entry("When arg is a int16", "a 16", int16(16), func(ctx context.Context, a int16) error { return nil }),
		//Entry("When arg is a int32", "a 32", int32(32), func(ctx context.Context, a int32) error { return nil }),
		//Entry("When arg is a int64", "a 64", int64(64), func(ctx context.Context, a int64) error { return nil }),
		//Entry("When arg is a float32", "a 3.2", float32(3.2), func(ctx context.Context, a float32) error { return nil }),
		//Entry("When arg is a float64", "a 6.4", float64(6.4), func(ctx context.Context, a float64) error { return nil }),
		//Entry("When arg is a []byte", "a bytes", []byte("bytes"), func(ctx context.Context, a []byte) error { return nil }),
	)
})

func TestScheme(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scheme Suite")
}
