package model

import (
	"context"
	"errors"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/testernetes/bdk/stepdef"
)

type StepRunner struct {
	Func   reflect.Value   `json:"-"`
	Args   []reflect.Value `json:"-"`
	Helper *stepdef.T
}

// Runs a Step Definition
// The result depends on the return type or panic. If the step:
// * returns nil: The step result is passed
// * returns err: The step result is unknown as the step itself failed to run
// * panics string: The step result is failed as string is a failure message typically from Gomega
// * panics any: The step result is unknown as the step itself failed to run
func (s *StepRunner) Run() (result stepdef.StepResult, err error) {
	ctx := s.Args[0].Interface().(context.Context)
	var ret []reflect.Value

	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		if r := recover(); r != nil {
			result.Err = errors.New(string(debug.Stack()))
			result.StartTime = startTime
			result.EndTime = endTime
			result.Result = stepdef.Unknown
			return
		}

		result.StartTime = startTime
		result.EndTime = endTime

		if errors.Is(ctx.Err(), context.Canceled) {
			result.Result = stepdef.Interrupted
			result.Messages = append(result.Messages, "Step canceled")
			return
		}

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.Result = stepdef.Timedout
			result.Messages = append(result.Messages, "Step timedout")
			return
		}

		var ok bool
		if err, ok = ret[0].Interface().(error); ok {
			result.Result = stepdef.Failed
			result.Messages = append(result.Messages, "Step failed")
			result.Err = err
			return
		}

		result.Err = nil
		result.Result = stepdef.Passed
		result.Messages = append(result.Messages, "Step passed")
	}()

	ret = s.Func.Call(s.Args)
	result = s.Helper.GetResult()
	return
}
