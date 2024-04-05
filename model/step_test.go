package model

import (
	"context"
	"fmt"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/stepdef"
)

var _ = Describe("Running Steps", func() {

	var basicGoodStep = stepdef.StepDefinition{
		Text: "a <text>",
		Function: func(ctx context.Context, s string) error {
			if len(s) > 1 {
				return nil
			}
			return fmt.Errorf("small string")
		},
	}

	Context("Running Basic Steps", Ordered, func() {
		scheme := &stepFunctions{}

		BeforeAll(func() {
			Expect(scheme.register(basicGoodStep)).Should(Succeed())
		})

		It("should run a basic good step", func() {
			stepDoc := &messages.Step{
				Text: "a word",
			}
			step := scheme.Eval(context.TODO(), stepDoc)
			Expect(step).ShouldNot(BeNil())

			step.Run()
			Expect(step.Execution.Result).Should(Equal(Passed))
			Expect(step.Execution.Err).ShouldNot(HaveOccurred())
			Expect(step.Execution.StartTime).Should(BeTemporally("<", step.Execution.EndTime))
		})
	})

})
