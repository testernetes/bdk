package model

import (
	"context"
	"errors"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

type StepRunner struct {
	Func reflect.Value   `json:"-"`
	Args []reflect.Value `json:"-"`
}

type StepResult struct {
	Result    Result    `json:"result"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Message   string    `json:"message,omitempty"`
	Err       error     `json:"error,omitempty"`
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
func (s *StepRunner) Run() (result StepResult, err error) {
	ctx := s.Args[0].Interface().(context.Context)
	var ret []reflect.Value

	defer func() {
		result.EndTime = time.Now()
		r := recover()

		if errors.Is(ctx.Err(), context.Canceled) {
			result.Result = Interrupted
			result.Message = "Step was interrupted"
			return
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.Result = Timedout
			result.Message = "Scenario timed out"
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
					result.Result = Timedout
					result.Message = message
					return
				}
				result.Result = Failed
				result.Message = message
				return
			}
			err = errors.New(string(debug.Stack()))
			result.Result = Unknown

			panic(r)
			return

		}

		if err, ok := ret[0].Interface().(error); ok {
			result.Result = Failed
			result.Message = err.Error()
			result.Err = err
			return
		}

		result.Result = Passed
		result.Message = "Step Ran Successfully"
	}()

	result.StartTime = time.Now()
	ret = s.Func.Call(s.Args)
	return
}
