package models

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"

	messages "github.com/cucumber/messages/go/v21"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

var (
	ErrNoAPIVersion = errors.New("Provided test case resource has an empty API Version")
	ErrNoKind       = errors.New("Provided test case resource has an empty Kind")
	ErrNoName       = errors.New("Provided test case resource has an empty Name")
)

type DocString struct {
	*messages.DocString
}

func (d *DocString) GetUnstructured() (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(d.DocString.Content), u)
	if err != nil {
		return u, err
	}

	if u.GetAPIVersion() == "" {
		return u, ErrNoAPIVersion
	}
	if u.GetKind() == "" {
		return u, ErrNoKind
	}
	if u.GetName() == "" {
		return u, ErrNoName
	}
	return u, nil
}

type Step struct {
	// Should these if templated by hydrated? yes, (maybe not if inject from previous step?)
	Location    *messages.Location       `json:"location"`
	Keyword     string                   `json:"keyword"`
	KeywordType messages.StepKeywordType `json:"keywordType,omitempty"`
	Text        string                   `json:"text"`
	DocString   *DocString               `json:"docString,omitempty"`
	DataTable   *messages.DataTable      `json:"dataTable,omitempty"`

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
func (s *Step) Run(ctx context.Context) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	args := append([]reflect.Value{reflect.ValueOf(ctx)}, s.Args...)

	var ret []reflect.Value

	defer func() {
		s.Execution.EndTime = time.Now()
		signal.Stop(c)
		r := recover()

		if errors.Is(ctx.Err(), context.Canceled) {
			s.Execution.Result = Interrupted
			s.Execution.Message = "Step was interrupted"
			return
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			s.Execution.Result = Timedout
			s.Execution.Message = "Scenario timed out"
			return
		}
		cancel()

		if r != nil {
			// Gomega panics with strings
			if message, ok := r.(string); ok {
				if strings.HasPrefix(message, "Timed out after") || strings.HasPrefix(message, "Context was cancelled after") {
					s.Execution.Result = Timedout
					s.Execution.Message = fmt.Sprintf("Timed out after %s", strings.Split(strings.SplitAfter(message, "after ")[1], "\n")[0])
					return
				}
				s.Execution.Result = Failed
				s.Execution.Message = message
				return
			}
			s.Execution.Result = Unknown
			s.Execution.Message = fmt.Sprintf("step paniced in an unexpected way: %s", r)
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

	go func() {
		select {
		case <-c:
			fmt.Printf("\nUser Interrupted, jumping to cleanup now. Press ^C again to skip cleanup.\n\n")
			cancel()
		case <-ctx.Done():
		}
	}()

	s.Execution.StartTime = time.Now()
	ret = s.Func.Call(args)
}

func NewStep(stepDoc *messages.Step, scheme *scheme) (*Step, error) {
	s := &Step{
		Location:    stepDoc.Location,
		Keyword:     stepDoc.Keyword,
		KeywordType: stepDoc.KeywordType,
		Text:        stepDoc.Text,
		DataTable:   stepDoc.DataTable,
	}

	if stepDoc.DocString != nil {
		s.DocString = &DocString{stepDoc.DocString}
	}

	err := scheme.StepDefFor(s)
	if err != nil {
		return nil, err
	}

	s.Execution.Result = Skipped

	return s, nil
}

func GenerateSteps(stepDocs []*messages.Step, scheme *scheme) ([]*Step, error) {
	var steps []*Step
	for _, stepDoc := range stepDocs {
		step, err := NewStep(stepDoc, scheme)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}
