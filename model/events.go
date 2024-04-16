package model

import (
	messages "github.com/cucumber/messages/go/v21"
	"github.com/testernetes/bdk/stepdef"
)

type EventType string

const (
	StartFeature   EventType = "StartFeature"
	FinishFeature  EventType = "FinishFeature"
	StartScenario  EventType = "StartScenario"
	FinishScenario EventType = "FinishScenario"
	StartStep      EventType = "StartStep"
	FinishStep     EventType = "FinishStep"
	InProgressStep EventType = "InProgressStep"
)

type Events chan Event

func (ch *Events) Close() {
	close(*ch)
}

func (ch *Events) StartFeature(feature *Feature) {
	*ch <- Event{Type: StartFeature, Feature: feature}
}

func (ch *Events) FinishFeature(feature *Feature) {
	*ch <- Event{Type: FinishFeature, Feature: feature}
}

func (ch *Events) StartScenario(scenario *Scenario) {
	*ch <- Event{Type: StartScenario, Scenario: scenario}
}

func (ch *Events) FinishScenario(scenario *Scenario) {
	*ch <- Event{Type: FinishScenario, Scenario: scenario}
}

func (ch *Events) StartStep(scenario *Scenario, step *messages.Step) {
	*ch <- Event{Type: StartStep, Scenario: scenario, Step: step}
}

func (ch *Events) InProgressStep(step *messages.Step, result stepdef.StepResult) {
	*ch <- Event{Type: InProgressStep, Step: step, StepResult: result}
}

func (ch *Events) FinishStep(scenario *Scenario, step *messages.Step) {
	*ch <- Event{Type: FinishStep, Scenario: scenario, Step: step}
}

type Event struct {
	Type EventType

	Feature    *Feature
	Scenario   *Scenario
	Step       *messages.Step
	StepResult stepdef.StepResult
}
