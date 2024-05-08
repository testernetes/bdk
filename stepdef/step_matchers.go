package stepdef

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/types"
	"sigs.k8s.io/yaml"
)

type Matcher struct {
	Name string
	Text string
	Help string
	Func any

	parameters []StringParameter
	re         *regexp.Regexp
}

func (m Matcher) GetExpression() *regexp.Regexp {
	return m.re
}

func (m Matcher) PrintHelp() string {
	return "TODO print parameters"
}

var Matchers = matchers{}

type matchers []Matcher

func (m *matchers) Register(matchers ...Matcher) {
	for _, matcher := range matchers {
		err := m.register(matcher)
		if err != nil {
			panic(err)
		}
	}
}

func (m *matchers) register(matcher Matcher) error {
	tFunc := reflect.TypeOf(matcher.Func)
	if tFunc.Kind() != reflect.Func {
		return errors.New(fmt.Sprintf("matcher '%s' has invalid Func", matcher.Text))
	}

	if tFunc.NumOut() != 1 {
		return errors.New(fmt.Sprintf("matcher '%s' does not return 1 value", matcher.Text))
	}
	if !tFunc.Out(0).Implements(reflect.TypeOf(new(types.GomegaMatcher)).Elem()) {
		return errors.New(fmt.Sprintf("matcher '%s' does not return a GomegaMatcher, it was %s", matcher.Text, tFunc.Out(0).String()))
	}

	// validate parameters and check matches regex capture groups and replace text with regex
	newText, params, err := StringParameters.SubstituteParameters(matcher.Text)
	if err != nil {
		return err
	}
	matcher.parameters = params

	// validate text and regex
	matcher.re, err = regexp.Compile(newText)
	if err != nil {
		return err
	}

	*m = append(*m, matcher)
	return nil
}

func (matchers *matchers) ParseMatcher(ctx context.Context, text string) (reflect.Value, error) {
	var matcher Matcher
	for _, m := range *matchers {
		if m.GetExpression().MatchString(text) {
			matcher = m
			break
		}
	}

	if matcher.Name == "" {
		return reflect.Value{}, fmt.Errorf("unrecognised matcher assertion: %s", text)
	}

	tFunc := reflect.TypeOf(matcher.Func)

	// if reflect Func has args fill them
	args := []reflect.Value{}
	submatches := matcher.GetExpression().FindSubmatch([]byte(text))
	for i := 0; i < tFunc.NumIn(); i++ {
		submatches = submatches[1:]
		if len(submatches) == 0 {
			break
		}
		v := submatches[0]

		// This is the last arg.
		// Split the last submatch into space delimited words and append to args
		// BeNumericallyThreashold
		// BeElementOf
		// ConsistOf
		if tFunc.IsVariadic() && i == tFunc.NumIn()-1 {
			words := getWords(v)
			for _, word := range words {
				args = append(args, reflect.ValueOf(word))
			}
			break
		}

		targetType := tFunc.In(i)
		if targetType.Kind() == reflect.String || targetType.Kind() == reflect.Interface {
			args = append(args, reflect.ValueOf(string(v)))
			continue
		}

		value := reflect.New(targetType)
		err := yaml.Unmarshal(v, value.Interface())
		if err != nil {
			return reflect.Value{}, err
		}
		args = append(args, reflect.Indirect(value))
	}
	ret := reflect.ValueOf(matcher.Func).Call(args)
	return ret[0], nil
}

func init() {
	// Order matters for matching due to anything wildcards
	Matchers.Register(matchers{
		{
			Name: "be-empty",
			Text: "be empty",
			Help: "Has a zero value",
			Func: gomega.BeEmpty,
		},
		{
			Name: "say",
			Text: "say {text}",
			Help: "Used with a PodSession such as logs or exec",
			Func: gbytes.Say,
		},
		{
			Name: "equal",
			Text: "equal {text}",
			Help: "Matches if strings are equal",
			Func: gomega.BeEquivalentTo,
		},
		{
			Name: "regex",
			Text: "match regex {text}",
			Help: "Matches if regex matches.",
			Func: gomega.MatchRegexp,
		},
		{
			Name: "len",
			Text: "have length {number}",
			Help: "Matches if the length equals expected.",
			Func: gomega.HaveLen,
		},
		{
			Name: "contains",
			Text: "contains {text}",
			Help: "Matches if the string contains the expected text.",
			Func: gomega.ContainSubstring,
		},
		{
			Name: "prefix",
			Text: "have prefix {text}",
			Help: "Matches if the string prefix has the expected text.",
			Func: gomega.HavePrefix,
		},
		{
			Name: "suffix",
			Text: "have suffix {text}",
			Help: "Matches if the string suffix has the expected text.",
			Func: gomega.HaveSuffix,
		},
		{
			Name: "numeric",
			Text: "be {comparator} {number}",
			Help: "Matches if the comparator between actual and expected passes.",
			Func: gomega.BeNumerically,
		},
		{
			Name: "numeric-threshold",
			Text: "be ~ {number} {number}",
			Help: "Matches if the actual and expected numbers are about equal within a threshold.",
			Func: gomega.BeNumerically,
		},
		{
			Name: "true",
			Text: "be true",
			Help: "Matches if true.",
			Func: gomega.BeTrue,
		},
		{
			Name: "false",
			Text: "be false",
			Help: "Matches if false.",
			Func: gomega.BeFalse,
		},
		{
			Name: "element-of",
			Text: "be an element of {text}",
			Help: "Matches if actual is one of the space delimited values",
			Func: gomega.BeElementOf,
		},
		{
			Name: "consist-of",
			Text: "consist of {text}",
			Help: "Matches if actual array consists of the expected values",
			Func: gomega.ConsistOf,
		},
	}...)
}

func getWords(in []byte) (out []interface{}) {
	scanner := bufio.NewScanner(bytes.NewReader(in))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		var word interface{}
		gomega.Expect(yaml.Unmarshal(scanner.Bytes(), &word)).Should(gomega.Succeed())
		out = append(out, word)
	}
	return
}
