package model

import (
	"context"
	"fmt"

	messages "github.com/cucumber/messages/go/v21"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testernetes/bdk/stepdef"
)

var _ = Describe("Running Scenarios", func() {

	var basicGoodStep = stepdef.StepDefinition{
		Text: "a <text>",
		Function: func(ctx context.Context, s string) error {
			if len(s) > 1 {
				return nil
			}
			return fmt.Errorf("small string")
		},
	}

	Context("Running Basic Scenarios", Ordered, func() {
		scheme := &stepFunctions{}

		BeforeAll(func() {
			Expect(scheme.register(basicGoodStep)).Should(Succeed())
		})

		It("should run a basic good step", func() {
			Skip("Add Kind and reenable")

			backgroundDoc := &messages.Background{
				Steps: []*messages.Step{
					{
						Text: "a word",
					},
				},
			}
			scenarioDoc := &messages.Scenario{
				Steps: []*messages.Step{
					{
						Text: "a word",
					},
				},
			}
			scenario, err := NewScenario(backgroundDoc, scenarioDoc, scheme)
			Expect(err).ShouldNot(HaveOccurred())

			scenario.Run(context.TODO())
			Expect(scenario.Steps[0].Execution.Result).Should(Equal(Passed))
			Expect(scenario.Steps[1].Execution.Result).Should(Equal(Passed))
		})
	})

})
