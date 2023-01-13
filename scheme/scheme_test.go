package scheme_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
)

var GoodStep = scheme.StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(ctx context.Context, s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
}

var StepWithoutContext = scheme.StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
}

var StepWithoutArgs = scheme.StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func() error {
		return nil
	},
}

var StepWithoutFunc = scheme.StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function:   "notafunction",
}

var StepTooFewArgs = scheme.StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(ctx context.Context) error {
		return nil
	},
}

var StepTooManyArgs = scheme.StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(ctx context.Context, s, b string) error {
		return nil
	},
}

var GoodDocStringStep = scheme.StepDefinition{
	Expression: regexp.MustCompile("find all (.*) in:"),
	Function: func(ctx context.Context, s string, doc *DocString) error {
		return nil
	},
}

var TooManyDocStringStep = scheme.StepDefinition{
	Expression: regexp.MustCompile("find all (.*) in:"),
	Function: func(ctx context.Context, s string, doc *DocString, doc2 *DocString) error {
		return nil
	},
}

var GoodDataTableStep = scheme.StepDefinition{
	Expression: regexp.MustCompile("find all (.*) in:"),
	Function: func(ctx context.Context, s string, doc *messages.DataTable) error {
		return nil
	},
}

var TooManyArgStep = scheme.StepDefinition{
	Expression: regexp.MustCompile("a (.*)"),
	Function: func(ctx context.Context, s string, doc *messages.DataTable) error {
		return nil
	},
}

var _ = Describe("Scheme", func() {
	Context("Adding Steps to Scheme", func() {
		DescribeTable("AddToScheme function",
			func(step scheme.StepDefinition, m GomegaMatcher) {
				s := scheme.Scheme{}
				Expect(scheme.AddToScheme(step)).Should(m)
			},
			Entry("should register a good step", GoodStep, Succeed()),
			Entry("should not register a  step without a context", StepWithoutContext, MatchError(ErrMustHaveContext)),
			Entry("should not register a  step without any arguments", StepWithoutArgs, MatchError(ErrMustHaveContext)),
			Entry("should not register a  step without a function", StepWithoutFunc, MatchError(ErrStepDefinitionMustHaveFunc)),
			Entry("should not register a  step which has too few args for the regular expression", StepTooFewArgs, MatchError(ErrTooFewArguments)),
			Entry("should not register a  step which has too many args for the regular expression", StepTooManyArgs, MatchError(ErrTooManyArguments)),
			Entry("should register a good step with DocString", GoodDocStringStep, Succeed()),
			Entry("should register a good step with DataTable", GoodDataTableStep, Succeed()),
			Entry("should not register a step with two DocString", TooManyDocStringStep, MatchError(ErrTooManyArguments)),
		)
	})
})

var _ = Describe("Hydrating a Step", func() {

	Context("Applying a Step Definition to a Step", func() {
		DescribeTable("Applying a matching step definition to a step",
			func(text string, arg interface{}, f interface{}) {
				var stepDef = scheme.StepDefinition{
					Expression: regexp.MustCompile("a (.*)"),
					Function:   f,
				}
				s := Scheme{}
				Expect(s.AddToScheme(stepDef)).Should(Succeed())

				stepFunc, stepArgs, err := s.StepDefFor(text)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(stepFunc).Should(Equal(reflect.ValueOf(f)))
				Expect(stepArgs).Should(HaveLen(1))
				Expect(stepArgs[0].Interface()).Should(Equal(arg))
			},
			Entry("When arg is a string", "a word", "word", func(ctx context.Context, a string) error { return nil }),
			Entry("When arg is a int", "a 47", 47, func(ctx context.Context, a int) error { return nil }),
			Entry("When arg is a int8", "a 8", int8(8), func(ctx context.Context, a int8) error { return nil }),
			Entry("When arg is a int16", "a 16", int16(16), func(ctx context.Context, a int16) error { return nil }),
			Entry("When arg is a int32", "a 32", int32(32), func(ctx context.Context, a int32) error { return nil }),
			Entry("When arg is a int64", "a 64", int64(64), func(ctx context.Context, a int64) error { return nil }),
			Entry("When arg is a float32", "a 3.2", float32(3.2), func(ctx context.Context, a float32) error { return nil }),
			Entry("When arg is a float64", "a 6.4", float64(6.4), func(ctx context.Context, a float64) error { return nil }),
			Entry("When arg is a []byte", "a bytes", []byte("bytes"), func(ctx context.Context, a []byte) error { return nil }),
		)

		DescribeTable("Applying a matching step definition with DocString or DataTable to a step",
			func(arg interface{}, f interface{}) {
				if ds, ok := arg.(*DocString); ok {
					step.DocString = ds
				}
				if dt, ok := arg.(*messages.DataTable); ok {
					step.DataTable = dt
				}
				var stepDef = scheme.StepDefinition{
					Expression: regexp.MustCompile("a (.*)"),
					Function:   f,
				}
				s := scheme.Scheme{}
				Expect(s.AddToScheme(stepDef)).Should(Succeed())

				stepFunc, stepArgs, err := s.StepDefFor("a blah")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(stepFunc).Should(Equal(reflect.ValueOf(f)))
				Expect(stepArgs).Should(HaveLen(2))
				Expect(stepArgs[1].Interface()).Should(Equal(arg))
			},
			Entry("When DocString",
				&DocString{&messages.DocString{Content: "helloworld"}},
				func(ctx context.Context, s string, doc *parameters.DocString) error { return nil },
			),
			Entry("When DataTable",
				&messages.DataTable{},
				func(ctx context.Context, s string, doc *messages.DataTable) error { return nil },
			),
		)

		It("should not apply for an invalid type", func() {
			var stepDef = scheme.StepDefinition{
				Expression: regexp.MustCompile("a (.*)"),
				Function:   func(ctx context.Context, m map[string]string) error { return nil },
			}
			scheme := Scheme{}
			Expect(scheme.Register(stepDef)).Should(MatchError(ErrUnsupportedArgumentType))
		})
	})
})

func TestScheme(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scheme Suite")
}
