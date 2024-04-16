package stepdef

import (
	"context"
	"fmt"
	"time"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/go-logr/logr"
	"github.com/testernetes/bdk/store"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type StepEvents interface {
	InProgressStep(*messages.Step, StepResult)
}

type StepResult struct {
	Result    Result    `json:"result"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Progress  float64
	Messages  []string `json:"messages,omitempty"`
	Err       error    `json:"error,omitempty"`
	Cleanup   []func() error
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

type T struct {
	Client    client.WithWatch
	Clientset kubernetes.Clientset
	Log       logr.Logger

	result StepResult
	events StepEvents
	step   *messages.Step
}

func NewT(ctx context.Context, sd StepDefinition, events StepEvents) *T {
	t := &T{
		events: events,
		step:   store.Load[*messages.Step](ctx, "step"),
		result: StepResult{
			StartTime: time.Now(),
		},
	}
	t.Log = log.FromContext(ctx).WithName(sd.Name).V(1)
	t.Client, _ = client.NewWithWatch(config.GetConfigOrDie(), client.Options{})
	t.Clientset = *kubernetes.NewForConfigOrDie(config.GetConfigOrDie())

	return t
}

func (t *T) notify() {
	t.events.InProgressStep(t.step, t.result)
}

func (t *T) Error(err error) {
	if err == nil {
		return
	}
	t.result.Result = Failed
	t.result.Err = err
	t.Log.Info("step is failing", "message", err.Error())
	t.notify()
}

func (t *T) GetResult() StepResult {
	return t.result
}

func (t *T) Errorf(format string, a ...any) {
	t.Error(fmt.Errorf(format, a...))
}

func (t *T) Cleanup(f func() error) {
	t.result.Cleanup = append(t.result.Cleanup, f)
	t.Log.Info("added post scenario cleanup", "func", f)
	t.notify()
}

func (t *T) SetProgress(percent float64) {
	t.result.Progress = percent
	t.Log.Info("progress set", "percent", percent)
	t.notify()
}

func (t *T) SetProgressGivenDuration(d time.Duration) {
	percent := float64(time.Since(t.result.StartTime)) / float64(d)
	t.result.Progress = percent
	t.Log.Info("progress set", "percent", percent)
	t.notify()
}

type IsRetryableFunc func(error) (bool, time.Duration)

func (t *T) WithRetry(ctx context.Context, f func() error, isRetryable IsRetryableFunc) (err error) {
	delay := time.After(0)
	for i := 0; ; i++ {
		t.Log.Info("retrying", "attempt", i)
		select {
		case <-ctx.Done():
			return
		case <-delay:
			err = f()
			retry, after := isRetryable(err)
			if !retry {
				return
			}
			delay = time.After(after)
		}
	}
}
