package steps

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/matchers"
	"github.com/testernetes/bdk/parameters"
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
	if contains(phrase, parameters.EventuallyPhrases) {
		return StringAsyncAssertion{AsyncAssertionTypeEventually, Eventually(f)}
	}
	if contains(phrase, parameters.ConsistentlyPhrases) {
		return StringAsyncAssertion{AsyncAssertionTypeConsistently, Consistently(f)}
	}
	panic("cannot determine if eventually or consistently")
}

func (assertion StringAsyncAssertion) WithContext(ctx context.Context, timeout string) StringAsyncAssertion {
	if timeout == "" {
		timeout = "1s"
	}
	d, err := time.ParseDuration(timeout)
	if err != nil {
		panic(fmt.Sprintf("cannot determine timeout: %w", err))
	}
	if assertion.asyncType == AsyncAssertionTypeEventually {
		ctx, _ = context.WithTimeout(ctx, d)
	}
	assertion.AsyncAssertion = assertion.AsyncAssertion.WithTimeout(d).WithContext(ctx)
	return assertion
}

func (assertion StringAsyncAssertion) WithArguments(args ...interface{}) StringAsyncAssertion {
	assertion.AsyncAssertion = assertion.AsyncAssertion.WithArguments(args...)
	return assertion
}

func (assertion StringAsyncAssertion) ShouldOrShouldNot(not, matcher string) bool {
	m := matchers.Matchers.GetMatcher(matcher)
	if not == "" {
		return assertion.AsyncAssertion.Should(m)
	}
	if not == "not" {
		return assertion.AsyncAssertion.ShouldNot(m)
	}
	panic("cannot determine if should or should not")
}

func contains(s string, a []string) bool {
	for i := range a {
		if a[i] == s {
			return true
		}
	}
	return false
}
