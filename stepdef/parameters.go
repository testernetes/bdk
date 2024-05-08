package stepdef

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	messages "github.com/cucumber/messages/go/v21"
)

const (
	RFC1123               = `([a-z0-9]+[-a-z0-9]*[a-z0-9])`
	DoubleQuoted          = `"([^"\\]*(?:\\.[^"\\]*)*)"`
	SingleQuoted          = `'([^'\\]*(?:\\.[^'\\]*)*)'`
	OneOrMultipleWords    = `([\w+\s*]+)`
	AsyncType             = OneOrMultipleWords
	exprDuration          = `((?:\d*\.?\d+h)?(?:\d*\.?\d+m)?(?:\d*\.?\d+s)?(?:\d*\.?\d+ms)?(?:\d*\.?\d+(?:us|µs))?(?:\d*\.?\d+ns)?)`
	exprShouldOrShouldNot = `((?:should\s?|to\s?)(?:not)?)`
	Anything              = `(.*)`
	exprArray             = Anything
	exprComparator        = `([=<>]{1,2})`
	exprNumber            = `(\d*\.?\d+)`
	exprMatcher           = Anything
	exprURLPath           = `([-a-zA-Z0-9()!@:%_\+.~#?&\/\/=]*)`
	exprURLScheme         = `(http|https)`
	exprPort              = `(\d{1,5})`
)

var CreateOptions = dataTableArgument{
	name:        "Create Options",
	description: "(optional) A table of additional client create options.",
	help: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#CreateOptions

		Create Options:
		| DryRun       | string  |
		| FieldOwner | string  |`,
	parser: ParseClientOptions,
}

var DeleteOptions = dataTableArgument{
	name:        "Delete Options",
	description: `(optional) A table of additional client delete options.`,
	help: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#DeleteOptions

		Delete Options:
		| DryRun             | string                         |
		| GracePeriodSeconds | number                         |
		| PropagationPolicy  | {Orphan|Background|Foreground} |`,
	parser: ParseClientOptions,
}

var DeleteAllOfOptions = dataTableArgument{
	name:        "Delete All Of Options",
	description: `(optional) A table of additional client delete all of options.`,
	help: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#DeleteAllOfOptions

		DeleteAllOf Options:
		| DryRun             | string                         |
		| GracePeriodSeconds | number                         |
		| PropagationPolicy  | (Orphan|Background|Foreground) |
		| Selector           | string                         |
		| Namespace          | string                         |
		| Limit              | number                         |`,
	parser: ParseClientOptions,
}

var ListOptions = dataTableArgument{
	name:        "List Options",
	description: `(optional) A table of additional client list options.`,
	help: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#ListOptions

		List Options:
		| Selector  | string |
		| Namespace | string |
		| Limit     | number |`,
	parser: ParseClientOptions,
}

var PatchOptions = dataTableArgument{
	name:        "Patch Options",
	description: `A table of additional client patch options.`,
	help: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#PatchOptions

		Patch Options:
		| DryRun       | string |
		| FieldOwner | string |
		| Force        | string |`,
	parser: ParseClientOptions,
}

var PodLogOptions = dataTableArgument{
	name:        "Pod Log Options",
	description: `(optional) A table of additional client pod log options.`,
	help: `https://pkg.go.dev/k8s.io/api/core/v1#PodLogOptions

		Pod Log Options:
		| container    | string  |
		| follow       | boolean |
		| previous     | boolean |
		| sinceSeconds | number  |
		| timestamps   | boolean |
		| tailLines    | number  |
		| limitBytes   | number  |`,
	parser: UnmarshalDataTable,
}

var ProxyGetOptions = dataTableArgument{
	name:        "Proxy Get Options",
	description: `(optional) A freeform table of additional query parameters to send with the request.`,
	help: `

		ProxyGet Options:
		| string | string |`,
	parser: func(ctx context.Context, table *messages.DataTable, targetType reflect.Type) (reflect.Value, error) {
		params := make(map[string]string)
		if table == nil {
			return reflect.ValueOf(params), nil
		}
		for _, row := range table.Rows {
			if len(row.Cells) < 2 {
				continue
			}
			opt := row.Cells[0].Value
			val := row.Cells[1].Value
			params[opt] = val
		}
		return reflect.ValueOf(params), nil
	},
}

var Patch = DocStringArgument{
	name:        "Patch",
	description: `A patch for a Kubernetes resource.`,
	help: `https://kubernetes.io/docs/reference/using-api/api-concepts/#patch-and-apply

		JSON Patch:
		"""application/json-patch+json
                [
                	{
                		"op" : "replace" ,
                		"path" : "/data/foo" ,
                		"value" : "nobar"
                	}
                ]
		"""

		Merge Patch
		"""application/merge-patch+json
          	{
	  	  "data": {
	  	    "foo":"nobar"
	  	  }
	  	}
		"""

		Strategic Patch
		"""application/strategic-merge-patch+json
          	{
	  	  "data": {
	  	    "foo":"nobar"
	  	  }
	  	}
		"""

		Server-Side Apply
		"""application/apply-patch+yaml
	        data:
	          foo: nobar
		"""`,
	parser: UnmarshalDocString,
}

var Manifest = DocStringArgument{
	name:        "Manifest",
	description: `A Kubernetes manifest.`,
	help: `https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/

		Can be yaml or json depending on the content type.`,
	parser: ParseDocStringToClientObject,
}

var Script = DocStringArgument{
	name:        "Script",
	description: `A script.`,
	help: `The script will run in the specified shell or if none is specified /bin/sh.
		Its outputs will be captured and can be asserted against in future steps.`,
	parser: UnmarshalDocString,
}

var MultiLineText = DocStringArgument{
	name:        "MultiLine Text",
	description: `A freeform DocString.`,
	help:        `Any multiline text.`,
	parser:      UnmarshalDocString,
}

var StringParameters = stringParameters{}

type stringParameters []StringParameter

func (sp *stringParameters) SubstituteParameters(step string) (expression string, params []StringParameter, err error) {
	re := regexp.MustCompile(`(?:{[^}]+})`)
	matches := re.FindAllStringSubmatch(step, -1)
	if len(matches) == 0 {
		return step, params, nil
	}

	expression = re.ReplaceAllStringFunc(step, func(name string) string {
		for _, p := range *sp {
			if p.Name() == name {
				params = append(params, p)
				return p.Expression()
			}
		}
		err = fmt.Errorf("no parameter registered for %s. %w", name, err)
		return ""
	})
	return
}

func (sp *stringParameters) Register(p ...StringParameter) {
	for _, next := range p {
		sp.register(next)
	}
}

func (sp *stringParameters) register(p StringParameter) {
	for _, existing := range *sp {
		if existing.Name() == p.Name() {
			panic("a parameter with the same name already exists")
		}
	}
	*sp = append(*sp, p)
}

func init() {
	StringParameters.Register(
		stringParameter{
			name:        "{filename}",
			expression:  Anything,
			description: `Path to a Kubernetes manifest.`,
			help: `https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/
	
			Can be yaml or json depending on the content type.`,
			parser: ParseFileToClientObject,
		},
		stringParameter{
			name:        "{jsonpath}",
			expression:  SingleQuoted,
			description: `A jsonpath to a field`,
			help:        `https://kubernetes.io/docs/reference/kubectl/jsonpath/`,
			parser:      StringParsers.Parse,
		},
		stringParameter{
			name:        "{assertion}",
			expression:  AsyncType,
			description: `An assertion that state should be achieved or maintained.`,
			help: fmt.Sprintf(`
		Eventually assertions can be made with: %q
		Consistently assertions can be made with: %q`, EventuallyPhrases, ConsistentlyPhrases),
			parser: StringParsers.Parse,
		},
		stringParameter{
			name:        "{reference}",
			expression:  RFC1123,
			description: `A short hand name for a resource.`,
			help: `https://kubernetes.io/docs/concepts/overview/working-with-objects/names/

		A resource can be assigned to this reference in a Context step and later referred to in an
		Action or Outcome step.

		The reference must a name that can be used as a DNS subdomain name as defined in RFC 1123.
		This is the same Kubernetes requirement for names, i.e. lowercase alphanumeric characters.`,
			parser: ParseClientObject,
		},
		stringParameter{
			name:        "{command}",
			expression:  DoubleQuoted,
			description: `The command to execute in the container.`,
			help: `https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/

		The command will run in a shell on the container and its outputs will be captured and can
		be asserted against in future steps.`,
			parser: StringParsers.Parse,
		},
		stringParameter{
			name:        "{container}",
			expression:  RFC1123,
			description: `The container name.`,
			help: `https://kubernetes.io/docs/concepts/overview/working-with-objects/names/

		The reference must a name that can be used as a DNS subdomain name as defined in RFC 1123.
		This is the same Kubernetes requirement for names, i.e. lowercase alphanumeric characters.`,
			parser: StringParsers.Parse,
		},
		stringParameter{
			name:        "{duration}",
			expression:  exprDuration,
			description: `Duration from when the step starts.`,
			help: `https://pkg.go.dev/time#ParseDuration

		A duration is a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5m" or "2h45m".
		Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`,
			parser: StringParsers.Parse,
		},
		stringParameter{
			name:        "{should|should not}",
			expression:  exprShouldOrShouldNot,
			description: `Used in conjunction with an assertion to assert that the actual matches the expected.`,
			help:        `To list available matchers run 'bdk matchers'.`,
			parser: func(ctx context.Context, s string, t reflect.Type) (reflect.Value, error) {
				if t.Kind() != reflect.Bool {
					return reflect.Value{}, errors.New("should or should not only supports parsing to bool")
				}
				s = strings.TrimSpace(s)
				if s == "should" || s == "to" {
					return reflect.ValueOf(true), nil
				}
				if s == "should not" || s == "to not" {
					return reflect.ValueOf(false), nil
				}
				return reflect.Value{}, fmt.Errorf("should or should not expression failed: received '%s'", s)
			},
		},
		stringParameter{
			name:        "{matcher}",
			expression:  Anything,
			description: `Used in conjunction with an assertion to assert that the actual matches the expected.`,
			help:        `To list available matchers run 'bdk matchers'.`,
			parser:      StringParsers.Parse,
		},
		stringParameter{
			name:        "{text}",
			expression:  Anything,
			description: `A freeform amount of text.`,
			help:        `This will match anything.`,
			parser:      StringParsers.Parse,
		},
		stringParameter{
			name:        "{number}",
			expression:  exprNumber,
			description: `A number.`,
			help:        `Can be decimal.`,
			parser:      StringParsers.Parse,
		},
		//stringParameter{
		//	name:        "{array}",
		//	expression:  exprArray,
		//	description: `A set of values.`,
		//	help:        `Must be space delimited.`,
		//	parser:      ParseArray,
		//},
		stringParameter{
			name:        "{comparator}",
			expression:  exprComparator,
			description: `a numeric comparator.`,
			help:        `One of ==, <, >, <=, >=.`,
			parser:      StringParsers.Parse,
		},
		stringParameter{
			name:        "{port}",
			expression:  exprPort,
			description: `Port.`,
			help:        `The port number to request. Acceptable range is 0 - 65536.`,
			parser: func(ctx context.Context, input string, targetType reflect.Type) (reflect.Value, error) {
				v, err := strconv.ParseInt(input, 10, 0)
				if err != nil {
					return reflect.Value{}, err
				}
				if v < 0 {
					return reflect.ValueOf(int(v)), errors.New("port must be a positive integer.")
				}
				if v > 65536 {
					return reflect.ValueOf(int(v)), errors.New("port must be less than 65536.")
				}
				if targetType.Kind() == reflect.String {
					return reflect.ValueOf(input), nil
				}
				if targetType.Kind() == reflect.Int {
					return reflect.ValueOf(int(v)), nil
				}
				return reflect.Value{}, fmt.Errorf("unsupported parameter type %v", targetType.Kind())
			},
		},
		stringParameter{
			name:        "{path}",
			expression:  exprURLPath,
			description: `The path of a URL.`,
			help:        `Anything that comes after port.`,
			parser:      StringParsers.Parse,
		},
		stringParameter{
			name:        "{scheme}",
			expression:  exprURLScheme,
			description: `The scheme of a URL.`,
			help:        `Must be either http or https.`,
			parser:      StringParsers.Parse,
		},
	)
}
