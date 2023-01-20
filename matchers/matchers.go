package matchers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/types"
	"github.com/testernetes/bdk/parameters"
	"github.com/testernetes/gkube"
	"sigs.k8s.io/yaml"
)

type Matcher struct {
	Name       string
	Text       string
	Help       string
	Parameters []parameters.StringParameter
	Expression *regexp.Regexp
	Func       interface{}
}

func (m Matcher) GetExpression() *regexp.Regexp {
	if m.Expression != nil {
		return m.Expression
	}
	expr := m.Text
	for _, p := range m.Parameters {
		expr = strings.ReplaceAll(expr, p.Text, p.Expression)
	}
	m.Expression = regexp.MustCompile(expr)
	return m.Expression
}

type matchers []Matcher

func (matchers matchers) GetMatcher(text string) types.GomegaMatcher {
	for _, m := range matchers {
		if m.GetExpression().MatchString(text) {

			mFunc := reflect.ValueOf(m.Func)
			if mFunc.Kind() != reflect.Func {
				panic(errors.New(fmt.Sprintf("matcher '%s' has invalid Func", m.Text)))
			}

			// panic if it doesn't return types.GomegaMatcher
			if mFunc.Type().NumOut() != 1 {
				panic(errors.New(fmt.Sprintf("matcher '%s' does not return 1 value", m.Text)))
			}
			if !mFunc.Type().Out(0).Implements(reflect.TypeOf(new(types.GomegaMatcher)).Elem()) {
				panic(errors.New(fmt.Sprintf("matcher '%s' does not return a GomegaMatcher, it was %s", m.Text, mFunc.Type().Out(0).String())))
			}

			// if reflect Func has args fill them
			args := []reflect.Value{}
			submatches := m.GetExpression().FindSubmatch([]byte(text))
			for i := 0; i < mFunc.Type().NumIn(); i++ {
				submatches = submatches[1:]
				if len(submatches) == 0 {
					break
				}
				v := submatches[0]

				// BeNumericallyThreashold
				// BeElementOf
				// ConsistOf
				if mFunc.Type().IsVariadic() && i == mFunc.Type().NumIn()-1 {
					// This is the last arg.
					// Split the last submatch into space delimited words and append to args
					words := getWords(v)
					for _, word := range words {
						args = append(args, reflect.ValueOf(word))
					}
					break
				}
				requiredKind := mFunc.Type().In(i).Kind()

				if requiredKind == reflect.String {
					args = append(args, reflect.ValueOf(string(v)))
					continue
				}

				var expected interface{}
				gomega.Expect(yaml.Unmarshal(v, &expected)).Should(gomega.Succeed())

				if requiredKind == reflect.Interface {
					args = append(args, reflect.ValueOf(expected))
					continue
				}
				if requiredKind == reflect.ValueOf(expected).Type().Kind() {
					args = append(args, reflect.ValueOf(expected))
					continue
				}
			}
			ret := mFunc.Call(args)
			return ret[0].Interface().(types.GomegaMatcher)
		}
	}
	panic(fmt.Sprintf("unrecognised assertion: %s", text))
}

var Matchers = matchers{}

func init() {
	// Order matters for matching due to anything wildcards
	Matchers = matchers{
		{
			Text: "jsonpath <jsonpath> <matcher>",
			Help: "Process the jsonpath",
			Func: func(j, text string) types.GomegaMatcher {
				m := Matchers.GetMatcher(text)
				return gkube.HaveJSONPath(j, m)
			},
			Parameters: []parameters.StringParameter{parameters.JSONPath, parameters.Matcher},
		},
		{
			Name: "be-empty",
			Text: "be empty",
			Help: "Has a zero value",
			Func: gomega.BeEmpty,
		},
		{
			Name:       "say",
			Text:       "say <text>",
			Help:       "Used with a PodSession such as logs or exec",
			Func:       gbytes.Say,
			Parameters: []parameters.StringParameter{parameters.Text},
		},
		{
			Name:       "equal",
			Text:       "equal <text>",
			Help:       "Matches if strings are equal",
			Func:       gomega.BeEquivalentTo,
			Parameters: []parameters.StringParameter{parameters.Text},
		},
		{
			Name:       "regex",
			Text:       "match regex <text>",
			Help:       "Matches if regex matches.",
			Func:       gomega.MatchRegexp,
			Parameters: []parameters.StringParameter{parameters.Text},
		},
		{
			Name:       "len",
			Text:       "have len <number>",
			Help:       "Matches if the length equals expected.",
			Func:       gomega.HaveLen,
			Parameters: []parameters.StringParameter{parameters.Number},
		},
		{
			Name:       "contains",
			Text:       "contains <text>",
			Help:       "Matches if the string contains the expected text.",
			Func:       gomega.ContainSubstring,
			Parameters: []parameters.StringParameter{parameters.Text},
		},
		{
			Name:       "prefix",
			Text:       "have prefix <text>",
			Help:       "Matches if the string prefix has the expected text.",
			Func:       gomega.HavePrefix,
			Parameters: []parameters.StringParameter{parameters.Text},
		},
		{
			Name:       "suffix",
			Text:       "have suffix <text>",
			Help:       "Matches if the string suffix has the expected text.",
			Func:       gomega.HaveSuffix,
			Parameters: []parameters.StringParameter{parameters.Text},
		},
		{
			Name:       "numeric",
			Text:       "be <comparator> <number>",
			Help:       "Matches if the comparator between actual and expected passes.",
			Func:       gomega.BeNumerically,
			Parameters: []parameters.StringParameter{parameters.Comparator, parameters.Number},
		},
		{
			Name:       "numeric-threshold",
			Text:       "be ~ <number> <number>",
			Help:       "Matches if the actual and expected numbers are about equal within a threshold.",
			Func:       gomega.BeNumerically,
			Parameters: []parameters.StringParameter{parameters.Number, parameters.Number},
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
			Name:       "element-of",
			Text:       "be an element of <text>",
			Help:       "Matches if actual is one of the space delimited values",
			Func:       gomega.BeElementOf,
			Parameters: []parameters.StringParameter{parameters.Array},
		},
		{
			Name:       "consist-of",
			Text:       "consist of <text>",
			Help:       "Matches if actual array consists of the expected values",
			Func:       gomega.ConsistOf,
			Parameters: []parameters.StringParameter{parameters.Array},
		},
	}
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
