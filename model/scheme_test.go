package model

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/parameters"
)

var GoodStep = StepDefinition{
	Text:       "a <text>",
	Parameters: []parameters.Parameter{parameters.Text},
	Function: func(ctx context.Context, s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
}

var StepWithoutContext = StepDefinition{
	Text: "a <text>",
	Function: func(s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
}

var StepWithoutArgs = StepDefinition{
	Text: "a <text>",
	Function: func() error {
		return nil
	},
}

var StepWithoutFunc = StepDefinition{
	Text:     "a <text>",
	Function: "notafunction",
}

var StepWithoutText = StepDefinition{}

var StepTooFewArgs = StepDefinition{
	Text:       "a <text>",
	Parameters: []parameters.Parameter{parameters.Text},
	Function: func(ctx context.Context) error {
		return nil
	},
}

var StepTooManyArgs = StepDefinition{
	Text:       "a <text>",
	Parameters: []parameters.Parameter{parameters.Text},
	Function: func(ctx context.Context, s, b string) error {
		return nil
	},
}

var _ = Describe("scheme", func() {
	Context("Adding Steps to scheme", func() {
		DescribeTable("AddTofunction",
			func(step StepDefinition, m types.GomegaMatcher) {
				s := 
			}
				Expect(s.AddTostep)).Should(m)
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

	Context("Find steps in the , func() {

	})
})
