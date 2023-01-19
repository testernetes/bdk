package scheme

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/testernetes/bdk/arguments"
	"github.com/testernetes/bdk/parameters"
)

// matchable errors
var (
	ErrUnmatchedStepArgumentNumber = errors.New("func received more arguments than expected")
	ErrCannotConvert               = errors.New("cannot convert argument")
	ErrUnsupportedArgumentType     = errors.New("unsupported argument type")
	ErrMustHaveContext             = errors.New("step function must have at least Context as the first argument")
	ErrStepDefinitionMustHaveFunc  = errors.New("must pass a function as the second argument to Register")
	ErrTooFewArguments             = errors.New("function has too few arguments for regular expression")
	ErrTooManyArguments            = errors.New("function has too many arguments for regular expression")
	ErrMustHaveErrReturn           = errors.New("function must only return error")
	ErrNoStepDefFound              = errors.New("cannot find a matching step definition")
)

type StepDefinition struct {
	Expression *regexp.Regexp
	Function   interface{}
	Name       string
	Text       string
	Examples   string
	Help       string
	Parameters []parameters.Parameter
}

func (s StepDefinition) GetExpression() *regexp.Regexp {
	if s.Expression != nil {
		return s.Expression
	}
	expr := s.Text
	for _, p := range s.Parameters {
		expr = strings.ReplaceAll(expr, p.Text, p.Expression)
	}
	s.Expression = regexp.MustCompile("^" + expr)
	//s.Expression = regexp.MustCompile(expr)
	return s.Expression
}

var Default = &Scheme{}

type Scheme struct {
	stepDefinitions []StepDefinition
}

func (s *Scheme) GetStepDefs() []StepDefinition {
	return s.stepDefinitions
}

func (s *Scheme) MustAddToScheme(sd StepDefinition) {
	err := s.AddToScheme(sd)
	if err != nil {
		panic(err)
	}
}

// TODO loop through all args and ensure they are valid types
// TODO ensure Parameters are all used
func (s *Scheme) AddToScheme(sd StepDefinition) error {
	inputs := sd.GetExpression().NumSubexp()

	stepFunc := reflect.ValueOf(sd.Function)
	if stepFunc.Kind() != reflect.Func {
		return ErrStepDefinitionMustHaveFunc
	}

	typ := stepFunc.Type()
	numIn := typ.NumIn()
	numOut := typ.NumOut()

	if numOut != 1 {
		return ErrMustHaveErrReturn
	}

	if !typ.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return ErrMustHaveErrReturn
	}

	if numIn < 1 {
		return ErrMustHaveContext
	}

	if !typ.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return ErrMustHaveContext
	}

	if numIn-1 < inputs {
		return ErrTooFewArguments
	}

	for i := 1; i < typ.NumIn(); i++ {
		param := typ.In(i)
		switch param.Kind() {
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.String, reflect.Float64, reflect.Float32:
			continue
		case reflect.Ptr:
			continue
			switch param.Elem().String() {
			case "arguments.DocString":
				continue
			case "arguments.DataTable":
				continue
				//default:
				//	return fmt.Errorf("%w: the argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Elem().String())
			}
		case reflect.Slice:
			switch param {
			case reflect.TypeOf([]byte(nil)):
				continue
			default:
				//return fmt.Errorf("%w: argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Kind())
				continue
			}
		default:
			//return fmt.Errorf("%w: argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Kind())
		}
	}

	// if it has exactly one extra arg and it is a DocString or DataTable
	if numIn-2 == inputs {
		//fmt.Printf("%s\n", typ.In(numIn-1).Kind().String())
		s.stepDefinitions = append(s.stepDefinitions, sd)
		return nil
		//if arg := typ.In(numIn - 1); arg.Kind() == reflect.Ptr {
		//	switch arg.Elem().String() {
		//	case "arguments.DocString":
		//		s.stepDefinitions = append(s.stepDefinitions, sd)
		//		return nil
		//	case "arguments.DataTable":
		//		s.stepDefinitions = append(s.stepDefinitions, sd)
		//		return nil
		//	}
		//}
		//return ErrTooManyArguments
	}

	if numIn-1 > inputs {
		return ErrTooManyArguments
	}

	s.stepDefinitions = append(s.stepDefinitions, sd)
	return nil
}

// Find a Step function which has a regular expression that matches the text input
// and same number of arguments, ignoring context and DocString or DataTable
// as they are not provided via the step text

// TODO change dt and ds to a single interface which converts to other objects
func (s *Scheme) StepDefFor(text string, dt *arguments.DataTable, ds *arguments.DocString) (reflect.Value, []reflect.Value, error) {
	var groups []string
	var stepDef StepDefinition
	var additionalStepArg int

	for _, sd := range s.stepDefinitions {
		additionalStepArg = 0
		if !sd.GetExpression().MatchString(text) {
			continue
		}

		// Get all capture groups
		groups = sd.GetExpression().FindStringSubmatch(text)

		// Check if there is an additional step argument
		if len(groups) == reflect.ValueOf(sd.Function).Type().NumIn()-1 {
			if sd.Parameters[len(sd.Parameters)-1].ParseDocString != nil {
				additionalStepArg = 1
			}
			if sd.Parameters[len(sd.Parameters)-1].ParseDataTable != nil {
				additionalStepArg = 1
			}
		}

		// the number of step capture groups (and additional step argument)
		// must match exactly the same number of parameters
		if len(groups)-1+additionalStepArg != len(sd.Parameters) {
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
	var stepArgs []reflect.Value

	// Parse regexp cature groups
	for i := 1; i < len(groups); i++ {
		targetType := stepFunc.Type().In(i)
		arg, err := stepDef.Parameters[i-1].ParseString(groups[i], targetType)
		if err != nil {
			return reflect.Value{}, []reflect.Value{}, fmt.Errorf("cannot parse parameter %s: %w", stepDef.Parameters[i-1].Text, err)
		}
		stepArgs = append(stepArgs, arg)
	}

	// Parse step argument if one exists
	if additionalStepArg == 1 {
		lastParam := len(stepDef.Parameters) - 1
		targetType := stepFunc.Type().In(lastParam + 1)
		if stepDef.Parameters[lastParam].ParseDocString != nil {
			arg, err := stepDef.Parameters[lastParam].ParseDocString(ds, targetType)
			if err != nil {
				return reflect.Value{}, []reflect.Value{}, fmt.Errorf("cannot parse parameter %s: %w", stepDef.Parameters[lastParam].Text, err)
			}
			stepArgs = append(stepArgs, arg)
		}
		if stepDef.Parameters[lastParam].ParseDataTable != nil {
			arg, err := stepDef.Parameters[lastParam].ParseDataTable(dt, targetType)
			if err != nil {
				return reflect.Value{}, []reflect.Value{}, fmt.Errorf("cannot parse parameter %s: %w", stepDef.Parameters[lastParam].Text, err)
			}
			stepArgs = append(stepArgs, arg)
		}
	}

	return stepFunc, stepArgs, nil
}
