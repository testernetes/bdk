package scheme

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
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
			switch param.Elem().String() {
			case "arguments.DocString":
				continue
			case "arguments.DataTable":
				continue
			default:
				return fmt.Errorf("%w: the argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Elem().String())
			}
		case reflect.Slice:
			switch param {
			case reflect.TypeOf([]byte(nil)):
				continue
			default:
				return fmt.Errorf("%w: argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Kind())
			}
		default:
			return fmt.Errorf("%w: argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Kind())
		}
	}

	// if it has exactly one extra arg and it is a DocString or DataTable
	if numIn-2 == inputs {
		if arg := typ.In(numIn - 1); arg.Kind() == reflect.Ptr {
			switch arg.Elem().String() {
			case "arguments.DocString":
				s.stepDefinitions = append(s.stepDefinitions, sd)
				return nil
			case "arguments.DataTable":
				s.stepDefinitions = append(s.stepDefinitions, sd)
				return nil
			}
		}
		return ErrTooManyArguments
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
	var input []string
	var fType reflect.Type

	var stepFunc reflect.Value
	var stepArgs []reflect.Value

	for _, sd := range s.stepDefinitions {
		if !sd.GetExpression().MatchString(text) {
			continue
		}

		input = sd.GetExpression().FindStringSubmatch(text)
		matchedInputs := len(input)

		f := reflect.ValueOf(sd.Function)
		fType = f.Type()
		fArgCount := fType.NumIn() // Ignoring context

		// Ignoring DocString or DataTable
		if lastParam := fType.In(fArgCount - 1); lastParam.Kind() == reflect.Ptr {
			if lastParam.Elem().String() == "arguments.DocString" {
				fArgCount--
			}
			if lastParam.Elem().String() == "arguments.DataTable" {
				fArgCount--
			}
		}

		if matchedInputs == fArgCount {
			stepFunc = f
			//fmt.Printf("Found step def for: %s == %s\n", text, sd.Text)
			break
		}
	}
	if !stepFunc.IsValid() {
		return stepFunc, stepArgs, errors.New(fmt.Sprintf("Can not find step definition for %s: %s", text, ErrNoStepDefFound))
	}

	// Build stepArgs from matched regexp values converting to their required type and storing as a reflect.Value
	// Ingoring first parameter context
	for i := 1; i < fType.NumIn(); i++ {
		param := fType.In(i)
		switch param.Kind() {
		case reflect.Int:
			v, err := strconv.ParseInt(input[i], 10, 0)
			if err != nil {
				return stepFunc, stepArgs, fmt.Errorf(`%w %d: "%s" to int: %s`, ErrCannotConvert, i, input[i], err)
			}
			stepArgs = append(stepArgs, reflect.ValueOf(int(v)))
		case reflect.Int64:
			v, err := strconv.ParseInt(input[i], 10, 64)
			if err != nil {
				return stepFunc, stepArgs, fmt.Errorf(`%w %d: "%s" to int64: %s`, ErrCannotConvert, i, input[i], err)
			}
			stepArgs = append(stepArgs, reflect.ValueOf(v))
		case reflect.Int32:
			v, err := strconv.ParseInt(input[i], 10, 32)
			if err != nil {
				return stepFunc, stepArgs, fmt.Errorf(`%w %d: "%s" to int32: %s`, ErrCannotConvert, i, input[i], err)
			}
			stepArgs = append(stepArgs, reflect.ValueOf(int32(v)))
		case reflect.Int16:
			v, err := strconv.ParseInt(input[i], 10, 16)
			if err != nil {
				return stepFunc, stepArgs, fmt.Errorf(`%w %d: "%s" to int16: %s`, ErrCannotConvert, i, input[i], err)
			}
			stepArgs = append(stepArgs, reflect.ValueOf(int16(v)))
		case reflect.Int8:
			v, err := strconv.ParseInt(input[i], 10, 8)
			if err != nil {
				return stepFunc, stepArgs, fmt.Errorf(`%w %d: "%s" to int8: %s`, ErrCannotConvert, i, input[i], err)
			}
			stepArgs = append(stepArgs, reflect.ValueOf(int8(v)))
		case reflect.String:
			stepArgs = append(stepArgs, reflect.ValueOf(input[i]))
		case reflect.Float64:
			v, err := strconv.ParseFloat(input[i], 64)
			if err != nil {
				return stepFunc, stepArgs, fmt.Errorf(`%w %d: "%s" to float64: %s`, ErrCannotConvert, i, input[i], err)
			}
			stepArgs = append(stepArgs, reflect.ValueOf(v))
		case reflect.Float32:
			v, err := strconv.ParseFloat(input[i], 32)
			if err != nil {
				return stepFunc, stepArgs, fmt.Errorf(`%w %d: "%s" to float32: %s`, ErrCannotConvert, i, input[i], err)
			}
			stepArgs = append(stepArgs, reflect.ValueOf(float32(v)))
		case reflect.Ptr:
			switch param.Elem().String() {
			case "arguments.DocString":
				stepArgs = append(stepArgs, reflect.ValueOf(ds))
			case "arguments.DataTable":
				stepArgs = append(stepArgs, reflect.ValueOf(dt))
			default:
				return stepFunc, stepArgs, fmt.Errorf("%w: the argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Elem().String())
			}
		case reflect.Slice:
			switch param {
			case reflect.TypeOf([]byte(nil)):
				stepArgs = append(stepArgs, reflect.ValueOf([]byte(input[i])))
			default:
				return stepFunc, stepArgs, fmt.Errorf("%w: the slice argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Kind())
			}
		default:
			return stepFunc, stepArgs, fmt.Errorf("%w: the argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Kind())
		}
	}

	return stepFunc, stepArgs, nil
}
