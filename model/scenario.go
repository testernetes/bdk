package model

import (
	"context"
	"fmt"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/testernetes/bdk/store"
	"github.com/testernetes/trackedclient"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Scenario struct {
	*messages.Scenario
	Background *messages.Background `json:"background"`
	event      chan *Scenario
}

func NewScenario(bkg *messages.Background, scn *messages.Scenario) (*Scenario, error) {
	if bkg == nil {
		bkg = &messages.Background{}
	}
	s := &Scenario{
		Scenario:   scn,
		Background: bkg,
	}
	return s, nil
}

func (s *Scenario) Run(ctx context.Context) bool {
	// add to ctx
	// * Helper
	tc, err := trackedclient.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		panic(err)
	}
	defer tc.DeleteAllTracked(ctx)

	ctx = store.NewStoreFor(ctx)

	// TODO add scenario / feature to ctx

	for _, step := range s.Background.Steps {
		err := s.runStep(ctx, step)
		if err != nil {
			return false
		}
	}

	for _, step := range s.Steps {
		err := s.runStep(ctx, step)
		if err != nil {
			return false
		}
	}
	return true
}

func (s *Scenario) runStep(ctx context.Context, step *messages.Step) error {
	// find a match from step definitions
	stepFunction := StepFunctions.Eval(ctx, step)
	if stepFunction == nil {
		return fmt.Errorf("no matching step defined")
	}
	stepFunction.Run()
	if stepFunction.Execution.Result != Passed {
		return fmt.Errorf(stepFunction.Execution.Message)
	}
	s.ping()
	return nil
}

func (s *Scenario) ping() {
	s.event <- s
}

type Background struct {
	Location    *messages.Location `json:"location"`
	Keyword     string             `json:"keyword"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Steps       []*Step            `json:"steps"`
}
