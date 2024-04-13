package model

import (
	"context"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/testernetes/bdk/store"
	"github.com/testernetes/trackedclient"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Scenario struct {
	*messages.Scenario
	Background  *messages.Background `json:"background"`
	StepResults map[*messages.Step]StepResult
}

func NewScenario(bkg *messages.Background, scn *messages.Scenario) (*Scenario, error) {
	if bkg == nil {
		bkg = &messages.Background{}
	}
	s := &Scenario{
		Scenario:    scn,
		Background:  bkg,
		StepResults: make(map[*messages.Step]StepResult),
	}
	return s, nil
}

func (s *Scenario) Run(ctx context.Context, events Events) error {
	events.StartScenario(s)

	tc, err := trackedclient.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		panic(err)
	}
	defer tc.DeleteAllTracked(ctx)

	ctx = store.NewStoreFor(ctx)

	store.Save(ctx, "scenario", s)
	store.Save(ctx, "clientWithWatch", tc.(client.WithWatch))
	//store.Save(ctx, "", tc) clientset

	for _, step := range s.Background.Steps {
		err := s.evalStep(ctx, events, step)
		if err != nil {
			return err
		}
	}

	for _, step := range s.Steps {
		err := s.evalStep(ctx, events, step)
		if err != nil {
			return err
		}
	}

	events.FinishScenario(s)
	return nil
}

func (s *Scenario) evalStep(ctx context.Context, events Events, step *messages.Step) (err error) {
	store.Save(ctx, "step", step)

	stepFunction, err := StepFunctions.Eval(ctx, step)
	if err != nil {
		return err
	}

	events.StartStep(s, step)
	defer events.FinishStep(s, step)

	s.StepResults[step], err = stepFunction.Run()
	return err
}
