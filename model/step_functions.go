package model

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/drone/envsubst"
	"github.com/testernetes/bdk/stepdef"
	"github.com/testernetes/bdk/steps"
	"github.com/testernetes/bdk/store"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	ErrUnmatchedStepArgumentNumber = errors.New("func received more arguments than expected")
	ErrCannotConvert               = errors.New("cannot convert argument")
	ErrUnsupportedArgumentType     = errors.New("unsupported argument type")
	ErrMustHaveContext             = errors.New("step function must have at least Context as the first argument")
	ErrMustHaveText                = errors.New("step must have text")
	ErrMustHaveName                = errors.New("step must have name")
	ErrMustHaveStepArg             = errors.New("StepArg cannot be nil, instead use stepdef.NoStepArg")
	ErrStepDefinitionMustHaveFunc  = errors.New("must pass a function as the second argument to Register")
	ErrTooFewArguments             = errors.New("function has too little arguments for regular expression")
	ErrTooManyArguments            = errors.New("function has too many arguments for regular expression")
	ErrMustHaveErrReturn           = errors.New("function must only return error")
	ErrFuncArgsMustMatchParams     = errors.New("cannot convert parameter into function argument")
)

func init() {
	StepFunctions.Register(
		steps.AResource,
		steps.AResourceFromFile,
		steps.APatch,
		steps.ICreate,
		steps.IDelete,
		steps.IEvict,
		steps.IExecInContainer,
		steps.IExecScriptInContainer,
		steps.IExecInDefaultContainer,
		steps.IExecScriptInDefaultContainer,
		steps.IGet,
		steps.IPatch,
		steps.IProxyGet,
		steps.AsyncAssertExec,
		steps.AsyncAssertExecWithTimeout,
		steps.AsyncAssertLog,
		steps.AsyncAssertLogWithTimeout,
		steps.AsyncAssert,
		steps.AsyncAssertWithTimeout,
		steps.AsyncAssertResp,
		steps.AsyncAssertRespWithTimeout,
	)
}

type stepFunction struct {
	stepdef.StepDefinition

	function   reflect.Value
	re         *regexp.Regexp
	parameters []stepdef.StringParameter
}

func (sf stepFunction) GetParameters() []stepdef.StringParameter {
	return sf.parameters
}

func (sf *stepFunction) Matches(step *messages.Step) bool {
	return sf.re.MatchString(step.Text)
}

// instanciate a stepdefinition given a step
func (sf *stepFunction) Eval(ctx context.Context, step *messages.Step, events *Events) (*StepRunner, error) {
	runner := &StepRunner{
		Func:   sf.function,
		Args:   []reflect.Value{reflect.ValueOf(ctx)},
		Helper: stepdef.NewT(ctx, sf.StepDefinition, events),
	}

	argOffset := 1 // ctx
	tFunc := sf.function.Type()

	targetType := tFunc.In(argOffset)
	if targetType == reflect.TypeOf(&stepdef.T{}) {
		runner.Args = append(runner.Args, reflect.ValueOf(runner.Helper))
		argOffset += 1
	}

	captureGroups := sf.re.FindStringSubmatch(step.Text)[1:]
	var matchedSoFar string
	for i, p := range sf.parameters {
		value := captureGroups[i]
		targetType := tFunc.In(argOffset + i)

		arg, err := p.Parse(ctx, value, targetType)
		if err != nil {
			fmt.Printf(matchedSoFar)
			return nil, fmt.Errorf("[%d]: %s => ???%s???: %w", i, value, targetType.String(), err)
		}
		matchedSoFar = fmt.Sprintf("%s[%d]: %s => %s\n", matchedSoFar, i, value, arg)
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

type stepFunctions []stepFunction

var StepFunctions = &stepFunctions{}

func (s *stepFunctions) Eval(ctx context.Context, step *messages.Step, events *Events) (*StepRunner, error) {
	text, err := variableSubstitution(ctx, step.Text)
	if err != nil {
		return nil, fmt.Errorf("step text: could not substitute variables: %w", err)
	}
	if text != "" {
		step.Text = text
	}

	if step.DocString != nil {
		ds, err := variableSubstitution(ctx, step.DocString.Content)
		if err != nil {
			return nil, fmt.Errorf("docstring: could not substitute variables: %w", err)
		}
		if ds != "" {
			step.DocString.Content = ds
		}
	}

	// TODO datatable replacement

	for _, sf := range *s {
		if sf.Matches(step) {
			log.FromContext(ctx).V(1).Info(sf.re.String())
			return sf.Eval(ctx, step, events)
		}
	}

	return nil, fmt.Errorf("could not find a matching step")
}

// return just an interface in future
func (sf *stepFunctions) Register(stepDefs ...stepdef.StepDefinition) {
	for _, s := range stepDefs {
		err := sf.register(s)
		if err != nil {
			fmt.Printf("failed to register StepDefinition %s\n", s.Name)
			panic(err)
		}
	}
}

func validateStepDefinition(sd stepdef.StepDefinition) error {
	if sd.Text == "" {
		return ErrMustHaveText
	}
	if sd.Name == "" {
		return ErrMustHaveName
	}
	if sd.Function == nil {
		return ErrStepDefinitionMustHaveFunc
	}
	if sd.StepArg == nil {
		return ErrMustHaveStepArg
	}

	vFunc := reflect.ValueOf(sd.Function)
	if vFunc.Kind() != reflect.Func {
		return ErrStepDefinitionMustHaveFunc
	}
	return validateFunction(vFunc)
}

func (sf *stepFunctions) register(sd stepdef.StepDefinition) (err error) {
	err = validateStepDefinition(sd)
	if err != nil {
		return err
	}

	tFunc := reflect.TypeOf(sd.Function)

	requiredNumParams := tFunc.NumIn() - 1 // context
	if sd.StepArg.StepArgType() != stepdef.NoStepArgType {
		requiredNumParams -= 1
	}
	if tFunc.NumIn() > 1 {
		if tFunc.In(1) == reflect.TypeOf(&stepdef.T{}) {
			requiredNumParams -= 1
		}
	}

	// validate parameters and check matches regex capture groups and replace text with regex
	newText, params, err := stepdef.StringParameters.SubstituteParameters(sd.Text)
	if err != nil {
		return err
	}

	if len(params) > requiredNumParams {
		return ErrTooFewArguments
	}

	if len(params) < requiredNumParams {
		fmt.Printf("%d string parameters detected and %d is needed to satisfy function", len(params), requiredNumParams)
		return ErrTooManyArguments
	}

	// validate text and regex
	re, err := regexp.Compile(newText)
	if err != nil {
		return err
	}

	s := stepFunction{
		StepDefinition: sd,

		function:   reflect.ValueOf(sd.Function),
		re:         re,
		parameters: params,
	}

	*sf = append(*sf, s)

	return nil
}

// Validates the following constraints hold for step function:
// * Function is a function
// * Function only returns an error
// * Function accepts context as its first arugment
func validateFunction(vFunc reflect.Value) error {
	tFunc := vFunc.Type()

	// returns an error
	if tFunc.NumOut() != 1 {
		return ErrMustHaveErrReturn
	}
	if !tFunc.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return ErrMustHaveErrReturn
	}

	// first parameter is context
	if tFunc.NumIn() < 1 {
		return ErrMustHaveContext
	}
	if !tFunc.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return ErrMustHaveContext
	}

	return nil
}

func variableSubstitution(ctx context.Context, s string) (string, error) {
	text, err := envsubst.EvalEnv(s)
	if err != nil {
		return text, err
	}
	return envsubst.Eval(text, func(key string) string {
		key = "scn-var-" + key
		return store.Load[string](ctx, key)
	})
}
