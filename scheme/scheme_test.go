package scheme_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
)

var GoodStep = scheme.StepDefinition{
	Text:       "a <text>",
	Parameters: []parameters.Parameter{parameters.Text},
	Function: func(ctx context.Context, s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
}

var StepWithoutContext = scheme.StepDefinition{
	Text: "a <text>",
	Function: func(s string) error {
		if len(s) > 1 {
			return nil
		}
		return fmt.Errorf("small string")
	},
}

var StepWithoutArgs = scheme.StepDefinition{
	Text: "a <text>",
	Function: func() error {
		return nil
	},
}

var StepWithoutFunc = scheme.StepDefinition{
	Text:     "a <text>",
	Function: "notafunction",
}

var StepWithoutText = scheme.StepDefinition{}

var StepTooFewArgs = scheme.StepDefinition{
	Text:       "a <text>",
	Parameters: []parameters.Parameter{parameters.Text},
	Function: func(ctx context.Context) error {
		return nil
	},
}

var StepTooManyArgs = scheme.StepDefinition{
	Text:       "a <text>",
	Parameters: []parameters.Parameter{parameters.Text},
	Function: func(ctx context.Context, s, b string) error {
		return nil
	},
}

var _ = Describe("Scheme", func() {
	Context("Adding Steps to Scheme", func() {
		DescribeTable("AddToScheme function",
			func(step scheme.StepDefinition, m types.GomegaMatcher) {
				s := scheme.Scheme{}
				Expect(s.AddToScheme(step)).Should(m)
			},
			Entry("should register a good step", GoodStep, Succeed()),
			Entry("should not register a step without a context", StepWithoutContext, MatchError(scheme.ErrMustHaveContext)),
			Entry("should not register a step without any arguments", StepWithoutArgs, MatchError(scheme.ErrMustHaveContext)),
			Entry("should not register a step without any text", StepWithoutText, MatchError(scheme.ErrMustHaveText)),
			Entry("should not register a step without a function", StepWithoutFunc, MatchError(scheme.ErrStepDefinitionMustHaveFunc)),
			Entry("should not register a step which has too few args for the regular expression", StepTooFewArgs, MatchError(scheme.ErrTooFewArguments)),
			Entry("should not register a step which has too many args for the regular expression", StepTooManyArgs, MatchError(scheme.ErrTooManyArguments)),
		)
	})

	Context("Find steps in the scheme", func() {

	})
})
