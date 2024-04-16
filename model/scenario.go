package model

import (
	"context"
	"errors"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
)

type Scenario struct {
	*messages.Scenario
	Background  *messages.Background `json:"background"`
	StepResults map[*messages.Step]stepdef.StepResult
}

func (s *Scenario) MarshalJSON() ([]byte, error) {
	return []byte(`"soon"`), nil
}

func NewScenario(bkg *messages.Background, scn *messages.Scenario) (*Scenario, error) {
	if bkg == nil {
		bkg = &messages.Background{}
	}
	s := &Scenario{
		Scenario:    scn,
		Background:  bkg,
		StepResults: make(map[*messages.Step]stepdef.StepResult),
	}
	return s, nil
}

func (s *Scenario) Run(ctx context.Context, events *Events) error {
	events.StartScenario(s)
	defer events.FinishScenario(s)

	ctx = store.NewStoreFor(ctx)

	store.Save(ctx, "scenario", s)

	var cleanups []func() error

	for _, step := range s.Background.Steps {
		res, err := s.evalStep(ctx, events, step)
		if err != nil {
			return err
		}
		s.StepResults[step] = res
		cleanups = append(cleanups, res.Cleanup...)
	}

	for _, step := range s.Steps {
		res, err := s.evalStep(ctx, events, step)
		if err != nil {
			return err
		}
		s.StepResults[step] = res
		cleanups = append(cleanups, res.Cleanup...)
	}

	var errs error
	for _, cleanup := range cleanups {
		errors.Join(errs, cleanup())
	}

	return errs
}

func (s *Scenario) evalStep(ctx context.Context, events *Events, step *messages.Step) (stepdef.StepResult, error) {
	store.Save(ctx, "step", step)

	stepFunction, err := StepFunctions.Eval(ctx, step, events)
	if err != nil {
		return stepdef.StepResult{}, err
	}

	events.StartStep(s, step)
	defer events.FinishStep(s, step)

	return stepFunction.Run()
}
