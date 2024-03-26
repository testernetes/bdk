package model

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	messages "github.com/cucumber/messages/go/v21"
)

//        Unspecified,
//        Context,
//        Action,
//        Outcome,
//        Conjunction,
//        Unknown

type Step struct {
	// Should these if templated by hydrated? yes, (maybe not if inject from previous step?)
	Location    *messages.Location       `json:"location"`
	Keyword     string                   `json:"keyword"`
	KeywordType messages.StepKeywordType `json:"keywordType,omitempty"`
	Text        string                   `json:"text"`
	DocString   *parameter.DocString     `json:"docString,omitempty"`
	DataTable   *parameter.DataTable     `json:"dataTable,omitempty"`

	// Step Definition
	Func reflect.Value   `json:"-"`
	Args []reflect.Value `json:"-"`

	// Step Result
	Execution StepExecution `json:"execution"`
}

type StepExecution struct {
	Result    Result    `json:"result"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Message   string    `json:"message,omitempty"`
	Err       error     `json:"error,omitempty"`
}

type stepErr struct {
	err error
}

func (e stepErr) Error() string {
	return e.err.Error()
}

func (e stepErr) MarshalJSON() ([]byte, error) {
	if e.err != nil {
		return []byte(fmt.Sprint(e)), nil
	}
	return []byte(""), nil
}

type Result int

const (
	Passed Result = iota
	Failed
	Skipped
	Interrupted
	Timedout
	Unknown
)

// Runs a Step Definition
// The result depends on the return type or panic. If the step:
// * returns nil: The step result is passed
// * returns err: The step result is unknown as the step itself failed to run
// * panics string: The step result is failed as string is a failure message typically from Gomega
// * panics any: The step result is unknown as the step itself failed to run
func (s *Step) Run() {
	ctx := s.Args[0].Interface().(context.Context)
	var ret []reflect.Value

	defer func() {
		s.Execution.EndTime = time.Now()
		r := recover()

		if errors.Is(ctx.Err(), context.Canceled) {
			//s.Execution.Result = Interrupted
			//s.Execution.Message = "Step was interrupted"
			return
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			s.Execution.Result = Timedout
			s.Execution.Message = "Scenario timed out"
			return
		}

		if r != nil {
			// Gomega panics with strings
			if message, ok := r.(string); ok {
				if strings.HasPrefix(message, "reflect:") {
					debug.PrintStack()
					panic(message)
				}
				if strings.HasPrefix(message, "Timed out after") || strings.HasPrefix(message, "Context was cancelled after") {
					s.Execution.Result = Timedout
					s.Execution.Message = message
					return
				}
				s.Execution.Result = Failed
				s.Execution.Message = message
				return
			}
			s.Execution.Result = Unknown
			s.Execution.Message = fmt.Sprintf("step panicked in an unexpected way: %v", r)
			s.Execution.Err = errors.New(string(debug.Stack()))
			return

		}

		if err, ok := ret[0].Interface().(error); ok {
			s.Execution.Result = Failed
			s.Execution.Message = err.Error()
			s.Execution.Err = stepErr{err}
			return
		}

		s.Execution.Result = Passed
		s.Execution.Message = "Step Ran Successfully"
	}()

	s.Execution.StartTime = time.Now()
	ret = s.Func.Call(s.Args)
}
