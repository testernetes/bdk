package stepdef

import (
	"errors"
	"time"

	"github.com/onsi/gomega/types"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	EventuallyPhrases   = []string{"", "within", "in less than", "in under", "in no more than"}
	ConsistentlyPhrases = []string{"for", "for at least", "for no less than"}
)

//type K8sExistMatcher struct {
//	expected client.Object
//}
//
//func (matcher *K8sExistMatcher) Match(actual interface{}) (success bool, err error) {
//	length, ok := lengthOf(actual)
//	if !ok {
//		return false, fmt.Errorf("BeEmpty matcher expects a string/array/map/channel/slice.  Got:\n%s", format.Object(actual, 1))
//	}
//
//	return length == 0, nil
//}
//
//func (matcher *K8sExistMatcher) FailureMessage(actual interface{}) (message string) {
//	return format.Message(actual, "to be empty")
//}
//
//func (matcher *K8sExistMatcher) NegatedFailureMessage(actual interface{}) (message string) {
//	return format.Message(actual, "not to be empty")
//}

type Assert func(bool, types.GomegaMatcher, any) (bool, error)

func Eventually(desiredMatch bool, matcher types.GomegaMatcher, actual any) (bool, error) {
	return assert(EventuallyAssertion, desiredMatch, matcher, actual)
}

func Consistently(desiredMatch bool, matcher types.GomegaMatcher, actual any) (bool, error) {
	return assert(ConsistentlyAssertion, desiredMatch, matcher, actual)
}

func assert(assertion assertion, desiredMatch bool, matcher types.GomegaMatcher, actual any) (bool, error) {
	matches, err := matcher.Match(actual)
	if err != nil {
		return false, err
	}

	if matches == desiredMatch {
		if assertion == EventuallyAssertion {
			return false, nil
		}
	}

	if matches != desiredMatch {
		if desiredMatch {
			err = errors.New(matcher.FailureMessage(actual))
		} else {
			err = errors.New(matcher.NegatedFailureMessage(actual))
		}
		if assertion == ConsistentlyAssertion {
			return false, err
		}
	}
	return true, err
}

type assertion string

const (
	EventuallyAssertion   assertion = "Eventually"
	ConsistentlyAssertion assertion = "Consistently"
)

func RetryK8sError(err error) (bool, time.Duration) {
	if err == nil {
		return false, 0
	}

	// Client issue
	if isRuntime(err) {
		return false, 0
	}

	// Generic connection issue
	// return generic errors so they can be retried (TODO maybe handle other http issues before returning)
	statusErr, ok := err.(*k8sErrors.StatusError)
	if !ok {
		return false, 0
	}

	if secondsToDelay, ok := k8sErrors.SuggestsClientDelay(statusErr); ok {
		return true, time.Duration(secondsToDelay) * time.Second
	}

	// if it can be retried
	reason := k8sErrors.ReasonForError(err)
	_, retryable := retryableReasons[reason]

	return retryable, 0
}

func isRuntime(err error) bool {
	return runtime.IsMissingKind(err) ||
		runtime.IsMissingVersion(err) ||
		runtime.IsNotRegisteredError(err) ||
		runtime.IsStrictDecodingError(err)
}

var retryableReasons = map[metav1.StatusReason]struct{}{
	metav1.StatusReasonNotFound:           {},
	metav1.StatusReasonServerTimeout:      {},
	metav1.StatusReasonTimeout:            {},
	metav1.StatusReasonTooManyRequests:    {},
	metav1.StatusReasonInternalError:      {},
	metav1.StatusReasonServiceUnavailable: {},
	metav1.StatusReasonConflict:           {},
}
