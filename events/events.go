package main

import "github.com/testernetes/bdk/model"

type EventType string

const (
	StartFeature   EventType = "StartFeature"
	FinishFeature  EventType = "FinishFeature"
	StartScenario  EventType = "StartScenario"
	FinishScenario EventType = "FinishScenario"
	StartStep      EventType = "StartStep"
	FinishStep     EventType = "FinishStep"
)

type Event struct {
	Type EventType

	Feature  *model.Feature
	Scenario *model.Scenario
	Step     *model.Step
}
