package model

import (
	"context"
	"fmt"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/bdk/scheme"
)

var _ = Describe("Running Steps", func() {

	var basicGoodStep = scheme.StepDefinition{
		Text:       "a <text>",
		Parameters: []parameters.Parameter{parameters.Text},
		Function: func(ctx context.Context, s string) error {
			if len(s) > 1 {
				return nil
			}
			return fmt.Errorf("small string")
		},
	}

	Context("Running Basic Steps", Ordered, func() {
		scheme := &scheme.Scheme{}

		BeforeAll(func() {
			Expect(scheme.AddToScheme(basicGoodStep)).Should(Succeed())
		})

		It("should run a basic good step", func() {
			stepDoc := &messages.Step{
				Text: "a word",
			}
			step, err := NewStep(stepDoc, scheme)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(step).ShouldNot(BeNil())

			step.Run(context.TODO())
			Expect(step.Execution.Result).Should(Equal(Passed))
			Expect(step.Execution.Err).ShouldNot(HaveOccurred())
			Expect(step.Execution.StartTime).Should(BeTemporally("<", step.Execution.EndTime))
		})
	})

})
