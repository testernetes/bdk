package model

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/drone/envsubst"
	"github.com/testernetes/bdk/contextutils"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/steps"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// matchable errors
var (
	ErrUnmatchedStepArgumentNumber = errors.New("func received more arguments than expected")
	ErrCannotConvert               = errors.New("cannot convert argument")
	ErrUnsupportedArgumentType     = errors.New("unsupported argument type")
	ErrMustHaveContext             = errors.New("step function must have at least Context as the first argument")
	ErrMustHaveText                = errors.New("step must have text")
	ErrMustHaveName                = errors.New("step must have name")
	ErrMustHaveStepArg             = errors.New("stepArg cannot be nil, instead pass parameters.NoStepArg")
	ErrStepDefinitionMustHaveFunc  = errors.New("must pass a function as the second argument to Register")
	ErrTooFewArguments             = errors.New("function has too few arguments for regular expression")
	ErrTooManyArguments            = errors.New("function has too many arguments for regular expression")
	ErrMustHaveErrReturn           = errors.New("function must only return error")
	ErrFuncArgsMustMatchParams     = errors.New("cannot convert parameter into function argument")
)

func init() {
	StepFunctions.Register(
		steps.AResource,
		steps.AResourceFromFile,
	)
}

type stepFunction struct {
	stepdef.StepDefinition

	function   reflect.Type
	re         *regexp.Regexp
	parameters []stepdef.StringParameter
}

var StepFunctions = &stepFunctions{}

type stepFunctions []stepFunction

func variableSubstitution(ctx context.Context, s string) (string, error) {
	text, err := envsubst.EvalEnv(s)
	if err != nil {
		return text, err
	}
	return envsubst.Eval(text, func(key string) string {
		return contextutils.LoadVariable(ctx, key)
	})
}

func (s *stepFunctions) Eval(ctx context.Context, step *messages.Step) *Step {
	text, err := variableSubstitution(ctx, step.Text)
	if err != nil {
		// TODO return err
		log.FromContext(ctx).Error(err, "step text: could not substitute variables")
	}
	if text != "" {
		step.Text = text
	}

	ds, err := variableSubstitution(ctx, step.DocString.Content)
	if err != nil {
		log.FromContext(ctx).Error(err, "docstring: could not substitute from environment variables")
	}
	if ds != "" {
		step.DocString.Content = ds
	}

	// TODO datatable replacement

	for _, sd := range *s {
		if stepRunner, match := sd.Matches(ctx, step); match {
			return stepRunner
		}
	}
	return nil
}

// return just an interface in future
func (sf *stepFunctions) Register(stepDefs ...stepdef.StepDefinition) {
	for _, s := range stepDefs {
		sf.register(s)
	}
}

func (sf *stepFunctions) register(input stepdef.StepDefinition) (err error) {
	s := stepFunction{
		StepDefinition: input,
	}
	if s.Text == "" {
		panic(ErrMustHaveText)
	}
	if s.Name == "" {
		panic(ErrMustHaveName)
	}
	if s.StepArg == nil {
		s.StepArg = stepdef.NoStepArg
	}

	otherIns := 1 // context.Context
	if s.StepArg.StepArgType() == stepdef.NoStepArgType {
		otherIns += 1
	}

	// validate function and matches regex capture groups
	s.function, err = processFunction(s.Function)
	if err != nil {
		return err
	}

	// validate parameters and check matches regex capture groups and replace text with regex
	newText, params, err := stepdef.StringParameters.SubstituteParameters(s.Text)
	if err != nil {
		return err
	}

	if s.function.NumIn()-otherIns < len(params) {
		return ErrTooFewArguments
	}

	if s.function.NumIn()-otherIns > len(params) {
		return ErrTooManyArguments
	}

	// validate text and regex
	s.re, err = regexp.Compile(newText)
	if err != nil {
		return err
	}

	*sf = append(*sf, s)

	return nil
}

// Validates the following constraints hold for step function:
// * Function is a function
// * Function only returns an error
// * Function accepts context as its first arugment
func processFunction(fn any) (reflect.Type, error) {
	if fn == nil {
		return nil, ErrStepDefinitionMustHaveFunc
	}
	vFunc := reflect.ValueOf(fn)

	if vFunc.Kind() != reflect.Func {
		return nil, ErrStepDefinitionMustHaveFunc
	}

	tFunc := vFunc.Type()

	// returns an error
	if tFunc.NumOut() != 1 {
		return nil, ErrMustHaveErrReturn
	}
	if !tFunc.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return nil, ErrMustHaveErrReturn
	}

	// first parameter is context
	if tFunc.NumIn() < 1 {
		return nil, ErrMustHaveContext
	}
	if !tFunc.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return nil, ErrMustHaveContext
	}

	return tFunc, nil
}

func (sf stepFunction) Matches(ctx context.Context, step *messages.Step) (*Step, bool) {
	if !sf.re.MatchString(step.Text) {
		return nil, false
	}
	stepRunner, err := sf.GetRunner(ctx, step)
	if err != nil {
		return stepRunner, true
	}
	return nil, true
}

// instanciate a stepdefinition given a step
func (sf stepFunction) GetRunner(ctx context.Context, step *messages.Step) (*Step, error) {
	runner := &Step{
		Location:    step.Location,
		Keyword:     step.Keyword,
		KeywordType: step.KeywordType,
		Text:        step.Text,
		Args:        []reflect.Value{reflect.ValueOf(ctx)},
	}

	captureGroups := sf.re.FindStringSubmatch(step.Text)[1:]

	// Parse regexp capture groups
	for i, stringValue := range captureGroups {
		targetType := sf.function.In(i + 1)
		p := sf.parameters[i]

		arg, err := p.Parse(ctx, stringValue, targetType)
		if err != nil {
			return nil, fmt.Errorf("cannot parse parameter %d: %w", i, err)
		}
		runner.Args = append(runner.Args, arg)
	}

	if sf.StepArg.StepArgType() != stepdef.NoStepArgType {
		targetType := sf.function.In(sf.function.NumIn() - 1)
		arg, err := sf.StepArg.Parse(step, targetType)
		if err != nil {
			return nil, fmt.Errorf("cannot parse step arg: %w", err)
		}
		runner.Args = append(runner.Args, arg)
	}

	runner.Execution.Result = Skipped

	return runner, nil
}
