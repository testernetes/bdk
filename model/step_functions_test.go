package model

import (
	"context"
	"fmt"
	"reflect"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
)

var GoodStep = stepdef.StepDefinition{
	Name: "good-step",
	Text: "a {text}",
	Function: func(ctx context.Context, s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
	StepArg: stepdef.NoStepArg,
}

var StepWithoutContext = stepdef.StepDefinition{
	Name: "no-ctx",
	Text: "a {text}",
	Function: func(s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
	StepArg: stepdef.NoStepArg,
}

var StepWithoutArgs = stepdef.StepDefinition{
	Name: "no-args",
	Text: "a {text}",
	Function: func() error {
		return nil
	},
	StepArg: stepdef.NoStepArg,
}

var StepWithoutFunc = stepdef.StepDefinition{
	Name:     "not-a-func",
	Text:     "a {text}",
	Function: "notafunction",
	StepArg:  stepdef.NoStepArg,
}

var StepWithoutText = stepdef.StepDefinition{}

var StepTooFewArgs = stepdef.StepDefinition{
	Name: "too-few-args",
	Text: "a {text}",
	Function: func(ctx context.Context) error {
		return nil
	},
	StepArg: stepdef.NoStepArg,
}

var StepTooManyArgs = stepdef.StepDefinition{
	Name: "too-many-args",
	Text: "a {text}",
	Function: func(ctx context.Context, s, b string) error {
		return nil
	},
	StepArg: stepdef.NoStepArg,
}

var _ = Describe("StepFunctions", func() {
	Context("Adding StepDefinitions to StepFunctions", func() {
		DescribeTable("AddTofunction",
			func(step stepdef.StepDefinition, m types.GomegaMatcher) {
				defer GinkgoRecover()
				var sf stepFunctions
				Expect(sf.register(step)).Should(m)
			},
			Entry("should register a good step", GoodStep, Succeed()),
			Entry("should not register a step without a context", StepWithoutContext, MatchError(ErrMustHaveContext)),
			Entry("should not register a step without any arguments", StepWithoutArgs, MatchError(ErrMustHaveContext)),
			Entry("should not register a step without any text", StepWithoutText, MatchError(ErrMustHaveText)),
			Entry("should not register a step without a function", StepWithoutFunc, MatchError(ErrStepDefinitionMustHaveFunc)),
			Entry("should not register a step which has too few args for the regular expression", StepTooFewArgs, MatchError(ErrTooFewArguments)),
			Entry("should not register a step which has too many args for the regular expression", StepTooManyArgs, MatchError(ErrTooManyArguments)),
		)
	})

	Context("Running a StepFunction", func() {
		It("should run", func() {
			var sf stepFunctions
			ctx := context.Background()
			Expect(sf.register(GoodStep)).Should(Succeed())
			Expect(sf).Should(HaveLen(1))

			Expect(sf[0].parameters).Should(HaveLen(1))
			Expect(sf[0].re).ShouldNot(BeNil())
			Expect(sf[0].function.Kind()).Should(Equal(reflect.Func))

			stepDef := &messages.Step{
				Text: "a step",
			}
			ctx = store.NewStoreFor(ctx)
			store.Save(ctx, "step", stepDef)
			stepFunc, err := sf.Eval(ctx, stepDef, nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(stepFunc).ShouldNot(BeNil())
			Expect(stepFunc.Args).Should(HaveLen(2))
			Expect(stepFunc.Func.Type().NumIn()).Should(Equal(2))
			ret := stepFunc.Func.Call(stepFunc.Args)
			Expect(ret).Should(HaveLen(1))
			Expect(ret[0].Interface()).Should(Succeed())
		})

	})
})
