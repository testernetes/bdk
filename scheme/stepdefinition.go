package scheme

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/testernetes/bdk/parameters"
)

// matchable errors
var (
	ErrUnmatchedStepArgumentNumber = errors.New("func received more arguments than expected")
	ErrCannotConvert               = errors.New("cannot convert argument")
	ErrUnsupportedArgumentType     = errors.New("unsupported argument type")
	ErrMustHaveContext             = errors.New("step function must have at least Context as the first argument")
	ErrMustHaveText                = errors.New("step must have text")
	ErrStepDefinitionMustHaveFunc  = errors.New("must pass a function as the second argument to Register")
	ErrTooFewArguments             = errors.New("function has too few arguments for regular expression")
	ErrTooManyArguments            = errors.New("function has too many arguments for regular expression")
	ErrMustHaveErrReturn           = errors.New("function must only return error")
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

func (sd StepDefinition) Matches(text string) bool {
	if !sd.GetExpression().MatchString(text) {
		return false
	}
	captureGroups := sd.GetExpression().NumSubexp()
	return sd.GetNumStringParams() == captureGroups
}

func (sd StepDefinition) GetNumStringParams() int {
	stepArg := 0
	if sd.hasStepArgument() {
		stepArg = 1
	}
	return len(sd.Parameters) - stepArg
}

func (s StepDefinition) GetExpression() *regexp.Regexp {
	if s.Expression != nil {
		return s.Expression
	}
	expr := s.Text
	for _, p := range s.Parameters {
		if param, ok := p.(parameters.StringParameter); ok {
			expr = strings.ReplaceAll(expr, param.GetText(), p.GetExpression())
		}
	}
	s.Expression = regexp.MustCompile("^" + expr)
	return s.Expression
}

func (sd StepDefinition) Valid() error {
	if sd.Text == "" {
		return ErrMustHaveText
	}

	err := sd.validParameters()
	if err != nil {
		return err
	}

	err = sd.validStepFunc()
	if err != nil {
		return err
	}

	return nil
}

func (sd StepDefinition) hasStepArgument() bool {
	if p := len(sd.Parameters) - 1; p >= 0 {
		lastParam := sd.Parameters[p]
		if _, ok := lastParam.(parameters.StringParameter); !ok {
			return true
		}
	}
	return false
}

// Validates the following constraints hold for step function:
// * Function is a function
// * Function only returns an error
// * Function accepts context as its first arugment
// * Function has correct number of parameters
func (sd StepDefinition) validStepFunc() error {
	vFunc := reflect.ValueOf(sd.Function)

	if vFunc.Kind() != reflect.Func {
		return ErrStepDefinitionMustHaveFunc
	}

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

	// has correct number of parameters, ignoring context
	if tFunc.NumIn()-1 < len(sd.Parameters) {
		return ErrTooFewArguments
	}

	if tFunc.NumIn()-1 > len(sd.Parameters) {
		return ErrTooManyArguments
	}

	return nil
}

// validates that all parameters are found in the step text
// if a step argument is used it should be the last parameter and have no text
func (sd StepDefinition) validParameters() error {
	captureGroups := sd.GetExpression().NumSubexp()

	if captureGroups != sd.GetNumStringParams() {
		return fmt.Errorf("Parameters and Step text inputs mismatch: Step text capture groups should be equal to the number of parameters. %d capture groups and %d parameters", captureGroups, sd.GetNumStringParams())
	}

	var err error
	for i, p := range sd.Parameters {
		if param, ok := p.(parameters.StringParameter); ok {
			if strings.Contains(sd.Text, param.GetText()) {
				continue
			}
			err = fmt.Errorf("Parameter %s not found in Step Definition text %w", param.GetText(), err)
		}
		if i != len(sd.Parameters)-1 {
			err = fmt.Errorf("Only the final Parameter can accept a DocString or DataTable. Parameter %d is not the last Parameter %w", i, err)
		}
	}
	return err
}
