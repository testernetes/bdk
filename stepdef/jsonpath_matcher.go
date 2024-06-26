package stepdef

import (
	"fmt"

	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type haveJSONPathMatcher struct {
	jsonpath string
	matcher  types.GomegaMatcher
	name     string
	value    interface{}
}

func NewHaveJSONPathMatcher(jpath string, matcher types.GomegaMatcher) *haveJSONPathMatcher {
	return &haveJSONPathMatcher{
		jsonpath: jpath,
		matcher:  matcher,
	}
}

func (m *haveJSONPathMatcher) Match(actual interface{}) (success bool, err error) {
	j := jsonpath.New("")
	if err := j.Parse(m.jsonpath); err != nil {
		return false, fmt.Errorf("JSON Path '%s' is invalid: %s", m.jsonpath, err.Error())
	}

	if o, ok := actual.(client.Object); ok {
		m.name = fmt.Sprintf("%T %s/%s", actual, o.GetNamespace(), o.GetName())
	} else {
		m.name = fmt.Sprintf("%T", actual)
	}

	var obj interface{}
	if u, ok := actual.(*unstructured.Unstructured); ok {
		obj = u.UnstructuredContent()
	} else {
		obj = actual
	}

	results, err := j.FindResults(obj)
	if err != nil {
		m.value = nil
		return false, nil
	}

	values := []interface{}{}
	for i := range results {
		for j := range results[i] {
			values = append(values, results[i][j].Interface())
		}
	}

	// Flatten values if single result
	if len(values) == 1 {
		m.value = values[0]
		return m.matcher.Match(values[0])
	}

	m.value = values
	return m.matcher.Match(values)
}

// FailureMessage returns a message comparing the full objects after an unexpected failure to match has occurred.
func (m *haveJSONPathMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%s at %s: %s", m.name, m.jsonpath, m.matcher.FailureMessage(m.value))
}

// NegatedFailureMessage returns a string comparing the full objects after an unexpected match has occurred.
func (m *haveJSONPathMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%s at %s: %s", m.name, m.jsonpath, m.matcher.NegatedFailureMessage(m.value))
}
