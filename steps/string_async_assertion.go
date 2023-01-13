package steps

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/types"
	. "github.com/testernetes/gkube"
	"gopkg.in/yaml.v2"
)

var (
	matcherHaveJSONPath = regexp.MustCompile(`^jsonpath '([^']+)' (.+)$`)
	matcherSay          = regexp.MustCompile(`^say (.*)$`)

	matcherEmpty     = regexp.MustCompile(`^be empty$`)
	matcherEqual     = regexp.MustCompile(`^equal (.*)$`)
	matcherElementOf = regexp.MustCompile(`^be an element of (.*)$`)
	matcherConsistOf = regexp.MustCompile(`^consist of (.*)$`)
	matcherNumeric   = regexp.MustCompile(`^(?:be )?([~=<>]{1,2}) (\d+)$`)
	matcherBool      = regexp.MustCompile(`^be (true|false)$`)
	matcherContain   = regexp.MustCompile(`^contain (.*)$`)
	matcherPrefix    = regexp.MustCompile(`^have prefix (.*)$`)
	matcherSuffix    = regexp.MustCompile(`^have suffix (.*)$`)
	matcherRegex     = regexp.MustCompile(`^match regex (.*)$`)
	matcherLen       = regexp.MustCompile(`^have length (\d)$`)
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
	switch phrase {
	case "", "within", "in less than", "in under", "in no more than":
		return StringAsyncAssertion{AsyncAssertionTypeEventually, Eventually(f)}
	case "for at least", "for no less than":
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
	m := getMatcher(matcher)
	if not == "" {
		return assertion.AsyncAssertion.Should(m)
	}
	if not == " not" {
		return assertion.AsyncAssertion.ShouldNot(m)
	}
	panic("cannot determine if should or should not")
}

func getMatcher(text string) types.GomegaMatcher {
	switch {
	case matcherHaveJSONPath.MatchString(text):
		opts := matcherHaveJSONPath.FindStringSubmatch(text)
		jsonpath := opts[1]
		return HaveJSONPath(jsonpath, getMatcher(opts[2]))

	case matcherSay.MatchString(text):
		expected := matcherSay.FindStringSubmatch(text)[1]
		return Say(expected)

	case matcherEmpty.MatchString(text):
		return BeEmpty()

	case matcherEqual.MatchString(text):
		expectedBytes := []byte(matcherEqual.FindStringSubmatch(text)[1])
		var expected interface{}
		Expect(yaml.Unmarshal(expectedBytes, &expected)).Should(Succeed())
		return BeEquivalentTo(expected)

	case matcherRegex.MatchString(text):
		expected := matcherRegex.FindStringSubmatch(text)[1]
		return MatchRegexp(string(expected))

	case matcherLen.MatchString(text):
		fields := matcherLen.FindStringSubmatch(text)
		expected, err := strconv.Atoi(string(fields[1]))
		Expect(err).ShouldNot(HaveOccurred(), "Length must be a positive integer")
		return HaveLen(expected)

	case matcherContain.MatchString(text):
		expected := matcherContain.FindStringSubmatch(text)[1]
		return ContainSubstring(string(expected))

	case matcherPrefix.MatchString(text):
		expected := matcherContain.FindStringSubmatch(text)[1]
		return HavePrefix(string(expected))

	case matcherSuffix.MatchString(text):
		expected := matcherContain.FindStringSubmatch(text)[1]
		return HaveSuffix(string(expected))

	case matcherNumeric.MatchString(text):
		fields := matcherNumeric.FindSubmatch([]byte(text))
		expectedBytes := fields[2]
		var expected interface{}
		Expect(yaml.Unmarshal(expectedBytes, &expected)).Should(Succeed())
		comparator := string(fields[1])
		return BeNumerically(comparator, expected)

	case matcherBool.MatchString(text):
		fields := matcherBool.FindStringSubmatch(string(text))
		if fields[1] == "true" {
			return BeTrue()
		}
		return BeFalse()

	case matcherElementOf.MatchString(text):
		fields := matcherElementOf.FindStringSubmatch(string(text))
		return BeElementOf(getWords(fields[1])...)

	case matcherConsistOf.MatchString(text):
		fields := matcherConsistOf.FindStringSubmatch(string(text))
		return ConsistOf(getWords(fields[1])...)

	default:
		panic(fmt.Sprintf("unrecognised assertion: %s", text))
	}
}

func getWords(in string) (out []interface{}) {
	scanner := bufio.NewScanner(strings.NewReader(in))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		var word interface{}
		Expect(yaml.Unmarshal(scanner.Bytes(), &word)).Should(Succeed())
		out = append(out, word)
	}
	return
}
