package steps

import (
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/stepdef"
)

type AsyncAssertionType uint

const (
	AsyncAssertionTypeEventually AsyncAssertionType = iota
	AsyncAssertionTypeConsistently
)

type StringAsyncAssertion struct {
	asyncType AsyncAssertionType
	types.AsyncAssertion
}

func NewStringAsyncAssertion(phrase string, f interface{}) StringAsyncAssertion {
	if contains(phrase, stepdef.EventuallyPhrases) {
		return StringAsyncAssertion{AsyncAssertionTypeEventually, Eventually(f)}
	}
	if contains(phrase, stepdef.ConsistentlyPhrases) {
		return StringAsyncAssertion{AsyncAssertionTypeConsistently, Consistently(f)}
	}
	panic("cannot determine if eventually or consistently")
}

func contains(s string, a []string) bool {
	for i := range a {
		if a[i] == s {
			return true
		}
	}
	return false
}
