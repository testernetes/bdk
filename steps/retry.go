package steps

import (
	"context"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func withRetry(ctx context.Context, f func() error) (err error) {
	delay := time.After(0)
	for {
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

func isRetryable(err error) (bool, time.Duration) {
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
