package scheme

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/testernetes/bdk/arguments"
	"github.com/testernetes/bdk/parameters"
)

// matchable errors
var (
	ErrNoStepDefFound = errors.New("cannot find a matching step definition")
)

var Default = &Scheme{}

type Scheme struct {
	stepDefinitions []StepDefinition
}

func (s *Scheme) AddToScheme(sd StepDefinition) error {
	err := sd.Valid()
	if err != nil {
		return err
	}
	s.stepDefinitions = append(s.stepDefinitions, sd)
	return nil
}

func (s *Scheme) MustAddToScheme(sd StepDefinition) {
	err := s.AddToScheme(sd)
	if err != nil {
		panic(err)
	}
}

func (s *Scheme) GetStepDefs() []StepDefinition {
	return s.stepDefinitions
}

// Find a Step function which has a regular expression that matches the text input
// and same number of arguments, ignoring context.
func (s *Scheme) StepDefFor(text string, dt *arguments.DataTable, ds *arguments.DocString) (reflect.Value, []reflect.Value, error) {
	var stepDef StepDefinition

	for _, sd := range s.stepDefinitions {
		if !sd.Matches(text) {
			continue
		}
		stepDef = sd
		//fmt.Printf("Found step def for: %s == %s\n", text, sd.Text)
		break
	}
	if stepDef.Function == nil {
		return reflect.Value{}, []reflect.Value{}, errors.New(fmt.Sprintf("cannot find step definition for %s: %s", text, ErrNoStepDefFound))
	}
	stepFunc := reflect.ValueOf(stepDef.Function)

	captureGroups := stepDef.GetExpression().FindStringSubmatch(text)[1:]

	// Parse regexp cature groups
	var stepArgs []reflect.Value
	for i, p := range stepDef.Parameters {
		targetType := stepFunc.Type().In(i)
		var arg reflect.Value
		var err error

		switch param := p.(type) {
		case parameters.StringParameter:
			arg, err = param.Parser(captureGroups[i], targetType)
		case parameters.DocStringParameter:
			arg, err = param.Parser(ds, targetType)
		case parameters.DataTableParameter:
			arg, err = param.Parser(dt, targetType)
		default:
			return stepFunc, stepArgs, fmt.Errorf("unknown parameter type %T", param)
		}
		if err != nil {
			return reflect.Value{}, []reflect.Value{}, fmt.Errorf("cannot parse parameter %d: %w", i, err)
		}
		stepArgs = append(stepArgs, arg)
	}
	return stepFunc, stepArgs, nil
}
