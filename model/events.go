package model

import (
	messages "github.com/cucumber/messages/go/v21"
)

type EventType string

const (
	StartFeature   EventType = "StartFeature"
	FinishFeature  EventType = "FinishFeature"
	StartScenario  EventType = "StartScenario"
	FinishScenario EventType = "FinishScenario"
	StartStep      EventType = "StartStep"
	FinishStep     EventType = "FinishStep"
)

type Events chan Event

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
	*ch <- Event{Type: StartStep, Step: step}
}

func (ch *Events) FinishStep(scenario *Scenario, step *messages.Step) {
	*ch <- Event{Type: FinishStep, Step: step}
}

type Event struct {
	Type EventType

	Feature  *Feature
	Scenario *Scenario
	Step     *messages.Step
}
