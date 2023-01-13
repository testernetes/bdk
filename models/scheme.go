package models

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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
	Parameters []Parameter
}

func (s StepDefinition) GetExpression() *regexp.Regexp {
	if s.Expression != nil {
		return s.Expression
	}
	expr := s.Text
	for _, p := range s.Parameters {
		expr = strings.ReplaceAll(expr, p.Text, p.Expression)
	}
	s.Expression = regexp.MustCompile(expr)
	return s.Expression
}

var Scheme = &scheme{}

type scheme struct {
	stepDefinitions []StepDefinition
}

func (s *scheme) GetStepDefs() []StepDefinition {
	return s.stepDefinitions
}

// TODO loop through all args and ensure they are valid types
func (s *scheme) Register(sd StepDefinition) error {
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
			case "models.DocString":
				continue
			case "messages.DataTable":
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

	// if it has exactly one extra arg assert that it is a DocString or DataTable
	if numIn-2 == inputs {
		if arg := typ.In(numIn - 1); arg.Kind() == reflect.Ptr {
			switch arg.Elem().String() {
			case "models.DocString":
				s.stepDefinitions = append(s.stepDefinitions, sd)
				return nil
			case "messages.DataTable":
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
func (s *scheme) StepDefFor(step *Step) error {
	var matched bool
	var input []string
	var fType reflect.Type

	for _, sd := range s.stepDefinitions {
		if !sd.GetExpression().MatchString(step.Text) {
			continue
		}

		input = sd.GetExpression().FindStringSubmatch(step.Text)
		matchedInputs := len(input)

		f := reflect.ValueOf(sd.Function)
		fType = f.Type()
		fArgCount := fType.NumIn() // Ignoring context

		// Ignoring DocString or DataTable
		if lastParam := fType.In(fArgCount - 1); lastParam.Kind() == reflect.Ptr {
			// lastParam.Elem() == reflect.TypeOf((*messages.DocString)(nil)).Elem()
			if lastParam.Elem().String() == "models.DocString" {
				fArgCount--
			}
			if lastParam.Elem().String() == "messages.DataTable" {
				fArgCount--
			}
		}

		if matchedInputs == fArgCount {
			step.Func = f
			matched = true
			break
		}
	}
	if !matched {
		return errors.New(fmt.Sprintf("Can not find step definition for %s: %s", step.Text, ErrNoStepDefFound))
	}
	fmt.Printf("Found step def for: %s == %s\n", step.Text, runtime.FuncForPC(step.Func.Pointer()).Name())

	// Build step.Args from matched regexp values converting to their required type and storing as a reflect.Value
	// Ingoring first parameter context
	for i := 1; i < fType.NumIn(); i++ {
		param := fType.In(i)
		switch param.Kind() {
		case reflect.Int:
			v, err := strconv.ParseInt(input[i], 10, 0)
			if err != nil {
				return fmt.Errorf(`%w %d: "%s" to int: %s`, ErrCannotConvert, i, input[i], err)
			}
			step.Args = append(step.Args, reflect.ValueOf(int(v)))
		case reflect.Int64:
			v, err := strconv.ParseInt(input[i], 10, 64)
			if err != nil {
				return fmt.Errorf(`%w %d: "%s" to int64: %s`, ErrCannotConvert, i, input[i], err)
			}
			step.Args = append(step.Args, reflect.ValueOf(v))
		case reflect.Int32:
			v, err := strconv.ParseInt(input[i], 10, 32)
			if err != nil {
				return fmt.Errorf(`%w %d: "%s" to int32: %s`, ErrCannotConvert, i, input[i], err)
			}
			step.Args = append(step.Args, reflect.ValueOf(int32(v)))
		case reflect.Int16:
			v, err := strconv.ParseInt(input[i], 10, 16)
			if err != nil {
				return fmt.Errorf(`%w %d: "%s" to int16: %s`, ErrCannotConvert, i, input[i], err)
			}
			step.Args = append(step.Args, reflect.ValueOf(int16(v)))
		case reflect.Int8:
			v, err := strconv.ParseInt(input[i], 10, 8)
			if err != nil {
				return fmt.Errorf(`%w %d: "%s" to int8: %s`, ErrCannotConvert, i, input[i], err)
			}
			step.Args = append(step.Args, reflect.ValueOf(int8(v)))
		case reflect.String:
			step.Args = append(step.Args, reflect.ValueOf(input[i]))
		case reflect.Float64:
			v, err := strconv.ParseFloat(input[i], 64)
			if err != nil {
				return fmt.Errorf(`%w %d: "%s" to float64: %s`, ErrCannotConvert, i, input[i], err)
			}
			step.Args = append(step.Args, reflect.ValueOf(v))
		case reflect.Float32:
			v, err := strconv.ParseFloat(input[i], 32)
			if err != nil {
				return fmt.Errorf(`%w %d: "%s" to float32: %s`, ErrCannotConvert, i, input[i], err)
			}
			step.Args = append(step.Args, reflect.ValueOf(float32(v)))
		case reflect.Ptr:
			switch param.Elem().String() {
			case "models.DocString":
				step.Args = append(step.Args, reflect.ValueOf(step.DocString))
			case "messages.DataTable":
				step.Args = append(step.Args, reflect.ValueOf(step.DataTable))
			default:
				return fmt.Errorf("%w: the argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Elem().String())
			}
		case reflect.Slice:
			switch param {
			case reflect.TypeOf([]byte(nil)):
				step.Args = append(step.Args, reflect.ValueOf([]byte(input[i])))
			default:
				return fmt.Errorf("%w: the slice argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Kind())
			}
		default:
			return fmt.Errorf("%w: the argument %d type %s is not supported", ErrUnsupportedArgumentType, i, param.Kind())
		}
	}

	return nil
}
