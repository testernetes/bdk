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
	"github.com/testernetes/bdk/parameters"
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

// return just an interface in future
func NewStepDefinition(name, text, help, examples string, fn any, stepArg parameters.StepArgParameter) (*StepDefinition, error) {
	if text == "" {
		return nil, ErrMustHaveText
	}
	if name == "" {
		return nil, ErrMustHaveName
	}
	if stepArg == nil {
		return nil, ErrMustHaveStepArg
	}

	// validate function and matches regex capture groups
	f, err := processFunction(fn)
	if err != nil {
		return nil, err
	}

	// validate parameters and check matches regex capture groups and replace text with regex
	newText, params, err := processParameters(text)
	if err != nil {
		return nil, err
	}

	acceptsDataTable := false
	acceptsDocString := false
	if stepArg != nil {
		if _, ok := stepArg.(*parameters.DocStringParameter); ok {
			acceptsDocString = true
		} else if _, ok := stepArg.(*parameters.DataTableParameter); ok {
			acceptsDocString = true
		} else {
			return nil, ErrUnsupportedArgumentType
		}
		params = append(params, stepArg)
	}

	if f.NumIn()-1 < len(params) {
		return nil, ErrTooFewArguments
	}

	if f.NumIn()-1 > len(params) {
		return nil, ErrTooManyArguments
	}

	for i := 0; i < len(params); i++ {
		if !params[i].ConvertsTo(f.In(i + 1).Kind()) {
			return nil, fmt.Errorf("%w, Parameter %d does not support %s", ErrFuncArgsMustMatchParams, i, f.In(i+1).Kind().String())
		}
	}

	// validate text and regex
	re, err := regexp.Compile(newText)
	if err != nil {
		return nil, err
	}

	sd := &StepDefinition{
		Name:     name,
		Text:     text,
		Examples: examples,
		Help:     help,
		Function: f,

		Parameters:       params,
		Expression:       re,
		AcceptsDataTable: acceptsDataTable,
		AcceptsDocString: acceptsDocString,
	}

	return sd, nil
}

type StepDefinition struct {
	Name     string
	Text     string
	Examples string
	Help     string
	Function reflect.Type

	Expression       *regexp.Regexp
	Parameters       []parameters.Parameter
	AcceptsDocString bool
	AcceptsDataTable bool
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

func processParameters(text string) (string, []parameters.Parameter, error) {
	params := []parameters.Parameter{}

	// I think this should be owned by the parameters list
	re := regexp.MustCompile(`(?:<\w+>)+`)
	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return text, params, nil
	}

	var err error
	newText := re.ReplaceAllStringFunc(text, func(text string) string {
		p, ok := parameters.Parameters[text]
		if !ok {
			err = fmt.Errorf("no parameter registered for %s. %w", text, err)
			return ""
		}
		params = append(params, p)
		return p.GetExpression()
	})
	return newText, params, err
}

func (sd StepDefinition) Matches(ctx context.Context, step *messages.Step) (*Step, bool) {
	if !sd.GetExpression().MatchString(step.Text) {
		return nil, false
	}
	stepRunner, err := sd.GetRunner(ctx, step)
	if err != nil {
		return stepRunner, true
	}
	return nil, true
}

func (s StepDefinition) GetRunner(ctx context.Context, step *messages.Step) (*Step, error) {
	runner := &Step{
		Location:    step.Location,
		Keyword:     step.Keyword,
		KeywordType: step.KeywordType,
		Text:        step.Text,
		Args:        []reflect.Value{reflect.ValueOf(ctx)},
	}

	captureGroups := s.GetExpression().FindStringSubmatch(step.Text)[1:]

	// Parse regexp capture groups
	for i, stringValue := range captureGroups {
		targetType := s.Function.In(i + 1)
		p := s.Parameters[i].(*parameters.StringParameter)

		arg, err := p.Parser(stringValue, targetType)
		if err != nil {
			return nil, fmt.Errorf("cannot parse parameter %d: %w", i, err)
		}
		runner.Args = append(runner.Args, arg)
	}

	targetType := s.Function.In(s.Function.NumIn() - 1)
	obj := reflect.New(targetType).Interface()

	if s.AcceptsDataTable {
		if targetType.AssignableTo(reflect.TypeOf((*messages.DataTable)(nil))) {
			runner.Args = append(runner.Args, reflect.ValueOf(step.DataTable))
		} else {
			(&parameters.DataTable{step.DataTable}).UnmarshalInto(obj)
			runner.Args = append(runner.Args, reflect.ValueOf(obj))
		}

	}

	if s.AcceptsDocString {
		if targetType.AssignableTo(reflect.TypeOf((*messages.DocString)(nil))) {
			runner.Args = append(runner.Args, reflect.ValueOf(step.DocString))
		} else {
			(&parameters.DocString{step.DocString}).UnmarshalInto(obj)
			runner.Args = append(runner.Args, reflect.ValueOf(obj))
		}
	}

	runner.Execution.Result = Skipped

	return runner, nil
}

func (s StepDefinition) GetExpression() *regexp.Regexp {
	return s.Expression
}

var Default = &StepDefinitions{}

type StepDefinitions []*StepDefinition

func (s *StepDefinitions) Eval(ctx context.Context, step *messages.Step) *Step {
	text, err := envsubst.EvalEnv(step.Text)
	if err != nil {
		log.FromContext(ctx).Error(err, "step text: could not substitute from environment variables")
	}
	text, err = envsubst.Eval(text, func(key string) string {
		return contextutils.LoadVariable(ctx, key)
	})
	if err != nil {
		log.FromContext(ctx).Error(err, "step text: could not substitute from step variables")
	}
	if text != "" {
		step.Text = text
	}

	ds, err := envsubst.EvalEnv(step.DocString.Content)
	if err != nil {
		log.FromContext(ctx).Error(err, "docstring: could not substitute from environment variables")
	}
	ds, err = envsubst.Eval(ds, func(key string) string {
		return contextutils.LoadVariable(ctx, key)
	})
	if err != nil {
		log.FromContext(ctx).Error(err, "docstring: could not substitute from step variables")
	}
	if ds != "" {
		step.DocString.Content = ds
	}

	// TODO datatable replacement
	//dt, err := envsubst.EvalEnv(step.DataTable.
	//if err != nil {
	//	log.FromContext(ctx).Error(err, "datatable: could not substitute from environment variables")
	//}
	//dt, err = envsubst.Eval(ds, func(key string) string {
	//	return contextutils.LoadVariable(ctx, key)
	//})
	//if err != nil {
	//	log.FromContext(ctx).Error(err, "docstring: could not substitute from step variables")
	//}
	//if dt != "" {
	//	step.DataTable = dt
	//}

	for _, sd := range *s {
		if stepRunner, match := sd.Matches(ctx, step); match {
			return stepRunner
		}
	}
	return nil
}
