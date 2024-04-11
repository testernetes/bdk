package stepdef

import (
	"errors"

	"github.com/onsi/gomega/types"
)

var (
	EventuallyPhrases   = []string{"", "within", "in less than", "in under", "in no more than"}
	ConsistentlyPhrases = []string{"for", "for at least", "for no less than"}
)

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
