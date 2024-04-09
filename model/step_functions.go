package model

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/drone/envsubst"
	"github.com/testernetes/bdk/formatters/utils"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/store"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

//func init() {
//	StepFunctions.Register(
//		steps.AResource,
//		steps.AResourceFromFile,
//	)
//}

type stepFunction struct {
	stepdef.StepDefinition

	function   reflect.Value
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
		return store.Load[string](ctx, key)
	})
}

func (s *stepFunctions) Print(stepName string) string {
	buf := bytes.NewBufferString("")
	for _, sf := range *s {
		if sf.Name == stepName {
			for _, p := range sf.parameters {
				// TODO do better printing
				text := p.Name()
				fmt.Fprintf(buf, utils.Examples("\n%s:\n"), text)
				fmt.Fprintf(buf, utils.Parameter(p.Description()))
				fmt.Fprintf(buf, utils.Parameter(p.Help()))
			}
		}
	}
	return buf.String()
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

	if step.DocString != nil {
		ds, err := variableSubstitution(ctx, step.DocString.Content)
		if err != nil {
			log.FromContext(ctx).Error(err, "docstring: could not substitute from environment variables")
		}
		if ds != "" {
			step.DocString.Content = ds
		}
	}

	// TODO datatable replacement

	for _, sd := range *s {
		if stepRunner, match := sd.Matches(ctx, step); match {
			return stepRunner
		} else {
			panic("2kjkj")
		}
	}
	return nil
}

// return just an interface in future
func (sf *stepFunctions) Register(stepDefs ...stepdef.StepDefinition) {
	for _, s := range stepDefs {
		err := sf.register(s)
		if err != nil {
			fmt.Printf("%+v", s)
			panic(err)
		}
	}
}

func (sf *stepFunctions) register(input stepdef.StepDefinition) (err error) {
	s := stepFunction{
		StepDefinition: input,
	}
	if s.Text == "" {
		return ErrMustHaveText
	}
	if s.Name == "" {
		return ErrMustHaveName
	}
	if s.StepArg == nil {
		s.StepArg = stepdef.NoStepArg
	}

	// validate function and matches regex capture groups
	s.function, err = processFunction(s.Function)
	if err != nil {
		return err
	}

	otherIns := 1 // context.Context
	if s.StepArg.StepArgType() != stepdef.NoStepArgType {
		otherIns += 1
	}

	tFunc := s.function.Type()
	if tFunc.NumIn() > 1 {
		switch tFunc.In(1) {
		case reflect.TypeOf((client.WithWatch)(nil)), reflect.TypeOf(kubernetes.Clientset{}):
			otherIns += 1
		}
	}

	// validate parameters and check matches regex capture groups and replace text with regex
	newText, params, err := stepdef.StringParameters.SubstituteParameters(s.Text)
	if err != nil {
		return err
	}

	if tFunc.NumIn()-otherIns < len(params) {
		return ErrTooFewArguments
	}

	if tFunc.NumIn()-otherIns > len(params) {
		return ErrTooManyArguments
	}

	s.parameters = params

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
func processFunction(fn any) (reflect.Value, error) {
	if fn == nil {
		return reflect.Value{}, ErrStepDefinitionMustHaveFunc
	}
	vFunc := reflect.ValueOf(fn)

	if vFunc.Kind() != reflect.Func {
		return vFunc, ErrStepDefinitionMustHaveFunc
	}

	tFunc := vFunc.Type()

	// returns an error
	if tFunc.NumOut() != 1 {
		return vFunc, ErrMustHaveErrReturn
	}
	if !tFunc.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return vFunc, ErrMustHaveErrReturn
	}

	// first parameter is context
	if tFunc.NumIn() < 1 {
		return vFunc, ErrMustHaveContext
	}
	if !tFunc.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return vFunc, ErrMustHaveContext
	}

	return vFunc, nil
}

func (sf stepFunction) Matches(ctx context.Context, step *messages.Step) (*Step, bool) {
	if !sf.re.MatchString(step.Text) {
		return nil, false
	}
	stepRunner, err := sf.GetRunner(ctx, step)
	if err != nil {
		panic("dfd")
	}
	return stepRunner, true
}

// instanciate a stepdefinition given a step
func (sf stepFunction) GetRunner(ctx context.Context, step *messages.Step) (*Step, error) {
	runner := &Step{
		Step: step,

		Func: sf.function,
		Args: []reflect.Value{reflect.ValueOf(ctx)},

		Execution: StepExecution{
			Result: Skipped,
		},
	}

	argOffset := 1
	tFunc := sf.function.Type()

	// TODO add the clients in
	targetType := tFunc.In(argOffset)
	if targetType == reflect.TypeOf((*client.WithWatch)(nil)) {
		client := store.Load[client.WithWatch](ctx, "client")
		runner.Args = append(runner.Args, reflect.ValueOf(client))
		argOffset += 1
	}

	captureGroups := sf.re.FindStringSubmatch(step.Text)[1:]
	for i, p := range sf.parameters {
		value := captureGroups[i]
		targetType := tFunc.In(argOffset + i)

		arg, err := p.Parse(ctx, value, targetType)
		if err != nil {
			return nil, fmt.Errorf("cannot parse parameter %d with value %s into type %s: %w", i, value, targetType.String(), err)
		}
		runner.Args = append(runner.Args, arg)
	}

	if sf.StepArg.StepArgType() != stepdef.NoStepArgType {
		targetType := tFunc.In(tFunc.NumIn() - 1)
		arg, err := sf.StepArg.Parse(ctx, step, targetType)
		if err != nil {
			return nil, fmt.Errorf("cannot parse step arg: %w", err)
		}
		runner.Args = append(runner.Args, arg)
	}

	return runner, nil
}
