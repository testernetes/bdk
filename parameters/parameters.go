package parameters

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"

	messages "github.com/cucumber/messages/go/v21"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	RFC1123               = `([a-z0-9]+[-a-z0-9]*[a-z0-9])`
	DoubleQuoted          = `"([^"\\]*(?:\\.[^"\\]*)*)"`
	SingleQuoted          = `'([^'\\]*(?:\\.[^'\\]*)*)'`
	OneOrMultipleWords    = `([\w+\s*]+)`
	AsyncType             = OneOrMultipleWords
	exprDuration          = `((?:\d*\.?\d+h)?(?:\d*\.?\d+m)?(?:\d*\.?\d+s)?(?:\d*\.?\d+ms)?(?:\d*\.?\d+(?:us|µs))?(?:\d*\.?\d+ns)?)`
	exprShouldOrShouldNot = `(?:should\s?|to\s?)(not)?`
	Anything              = `(.*)`
	exprArray             = Anything
	exprComparator        = `([=<>]{1,2})`
	exprNumber            = `(\d*\.?\d+)`
	exprMatcher           = Anything
	exprURLPath           = `([-a-zA-Z0-9()!@:%_\+.~#?&\/\/=]*)`
	exprURLScheme         = `(http|https)`
	exprPort              = `(\d{1,5})`
)

var (
	EventuallyPhrases   = []string{"", "within", "in less than", "in under", "in no more than"}
	ConsistentlyPhrases = []string{"for at least", "for no less than"}
)

var CreateOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Slice},
		ShortHelp: `(optional) A table of additional client create options.`,
		LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#CreateOptions

		Create Options:
		| DryRun       | All     |
		| FieldManager | string  |`,
	},
	Parser: ParseDataTable,
}

var DeleteOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Slice},
		ShortHelp: `(optional) A table of additional client delete options.`,
		LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#DeleteOptions

		Delete Options:
		| DryRun             | All                            |
		| GracePeriodSeconds | number                         |
		| PropagationPolicy  | (Orphan|Background|Foreground) |`,
	},
	Parser: ParseDataTable,
}

var PatchOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Slice},
		ShortHelp: `A table of additional client patch options.`,
		LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#PatchOptions

		Patch Options:
		| DryRun       | All     |
		| FieldManager | string  |
		| Force        | boolean |

		| patch         | string (required) |
		| field manager | string            |`,
		// TODO
	},
	Parser: ParseDataTable, // might need a custom parser
}

var PodLogOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Ptr},
		ShortHelp: `(optional) A table of additional client pod log options.`,
		LongHelp: `https://pkg.go.dev/k8s.io/api/core/v1#PodLogOptions

		Pod Log Options:
		| container    | string  |
		| follow       | boolean |
		| previous     | boolean |
		| sinceSeconds | number  |
		| timestamps   | boolean |
		| tailLines    | number  |
		| limitBytes   | number  |`,
	},
	Parser: ParseDataTable,
}

var ProxyGetOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Map},
		ShortHelp: `(optional) A freeform table of additional query parameters to send with the request.`,
		LongHelp: `

		ProxyGet Options:
		| string | string |`,
	},
	Parser: func(table *messages.DataTable, targetType reflect.Type) (reflect.Value, error) {
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

var Manifest = DocStringParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Ptr},
		ShortHelp: `A Kubernetes manifest.`,
		LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/

		Can be yaml or json depending on the content type.`,
	},
	Parser: ParseDocString,
}

var Script = DocStringParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.String},
		ShortHelp: `A script.`,
		LongHelp: `The script will run in the specified shell or if none is specified /bin/sh.
		Its outputs will be captured and can be asserted against in future steps.`,
	},
	Parser: ParseDocString,
}

var MultiLineText = DocStringParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.String},
		ShortHelp: `A freeform DocString.`,
		LongHelp:  `Any multiline text.`,
	},
	Parser: ParseDocString,
}

var Parameters = map[string]Parameter{
	"<filename>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.Ptr},
			Expression: Anything,
			ShortHelp:  `Path to a Kubernetes manifest.`,
			LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/
	
			Can be yaml or json depending on the content type.`,
		},
		Parser: func(path string, targetType reflect.Type) (reflect.Value, error) {
			manifest, err := ioutil.ReadFile(path)
			if err != nil {
				return reflect.Value{}, err
			}

			u := &unstructured.Unstructured{}
			err = yaml.UnmarshalStrict(manifest, u)
			if err != nil {
				return reflect.Value{}, err
			}

			return reflect.ValueOf(u), nil
		},
	},
	"<assertion>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: AsyncType,
			ShortHelp:  `An assertion that state should be achieved or maintained.`,
			LongHelp: fmt.Sprintf(`
		Eventually assertions can be made with: %q
		Consistently assertions can be made with: %q`, EventuallyPhrases, ConsistentlyPhrases),
		},
		Parser: ParseString,
	},
	"<reference>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: RFC1123,
			ShortHelp:  `A short hand name for a resource.`,
			LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
		
		A resource can be assigned to this reference in a Context step and later referred to in an
		Action or Outcome step.
		
		The reference must a name that can be used as a DNS subdomain name as defined in RFC 1123.
		This is the same Kubernetes requirement for names, i.e. lowercase alphanumeric characters.`,
		},
		Parser: ParseString,
	},
	"<command>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: DoubleQuoted,
			ShortHelp:  `The command to execute in the container.`,
			LongHelp: `https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/

		The command will run in a shell on the container and its outputs will be captured and can
		be asserted against in future steps.`,
		},
		Parser: ParseString,
	},
	"<container>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: RFC1123,
			ShortHelp:  `The container name.`,
			LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
		
		The reference must a name that can be used as a DNS subdomain name as defined in RFC 1123.
		This is the same Kubernetes requirement for names, i.e. lowercase alphanumeric characters.`,
		},
		Parser: ParseString,
	},
	"<duration>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.Int64},
			Expression: exprDuration,
			ShortHelp:  `Duration from when the step starts.`,
			LongHelp: `https://pkg.go.dev/time#ParseDuration
		
		A duration is a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5m" or "2h45m".
		Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`,
		},
		Parser: func(duration string, targetType reflect.Type) (reflect.Value, error) {
			if duration == "" {
				duration = "1s"
			}
			d, err := time.ParseDuration(duration)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot determine duration: %w", err)
			}
			return reflect.ValueOf(d), nil
		},
	},
	"<jsonpath>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: SingleQuoted,
			ShortHelp:  `A JSON Path to a field in the referenced resource.`,
			LongHelp: `https://kubernetes.io/docs/reference/kubectl/jsonpath/
		
		e.g. '{.metadata.name}'.`,
		},
		Parser: ParseString,
	},
	"<matcher>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: Anything,
			ShortHelp:  `Used in conjunction with an assertion to assert that the actual matches the expected.`,
			LongHelp:   `To list available matchers run 'bdk matchers'.`,
		},
		Text:   "<matcher>",
		Parser: ParseString,
	},
	"<text>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: Anything,
			ShortHelp:  `A freeform amount of text.`,
			LongHelp:   `This will match anything.`,
		},
		Text:   "<text>",
		Parser: ParseString,
	},
	"<number>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64},
			Expression: exprNumber,
			ShortHelp:  `A number.`,
			LongHelp:   `Can be decimal.`,
		},
		Text:   "<number>",
		Parser: ParseNumber,
	},
	"<array>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.Slice},
			Expression: exprArray,
			ShortHelp:  `A set of values.`,
			LongHelp:   `Must be space delimited.`,
		},
		Text:   "<array>",
		Parser: ParseArray,
	},
	"<comparator>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: exprComparator,
			ShortHelp:  `a numeric comparator.`,
			LongHelp:   `One of ==, <, >, <=, >=.`,
		},
		Text:   "<comparator>",
		Parser: ParseString,
	},
	"(should|should not)": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: exprShouldOrShouldNot,
			ShortHelp:  `A positive or negative assertion.`,
			LongHelp:   `"to" can also be used instead of "should".`,
		},
		Text:   "(should|should not)",
		Parser: ParseString,
	},
	"<path>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.Int, reflect.String},
			Expression: exprPort,
			ShortHelp:  `Port.`,
			LongHelp:   `The port number to request. Acceptable range is 0 - 65536.`,
		},
		Text: "<port>",
		Parser: func(input string, targetType reflect.Type) (reflect.Value, error) {
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
	"<path>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: exprURLPath,
			ShortHelp:  `The path of a URL.`,
			LongHelp:   `Anything that comes after port.`,
		},
		Text:   "<path>",
		Parser: ParseString,
	},
	"<scheme>": StringParameter{
		BaseParameter: BaseParameter{
			Kinds:      []reflect.Kind{reflect.String},
			Expression: exprURLScheme,
			ShortHelp:  `The scheme of a URL.`,
			LongHelp:   `Must be either http or https.`,
		},
		Text:   "<scheme>",
		Parser: ParseString,
	},
}

//var OutWriter = StringParameter{
//	BaseParameter: BaseParameter{
//		Expression: RFC1123,
//		ShortHelp:  `A writer`,
//		LongHelp: `Where to write an out stream
//		Options:
//		  - stdout
//		  - stdin
//		  - file:<path>
//		  - null
//		`,
//	},
//	Text:   "<out>",
//	Parser: ParseWriter,
//}
//
//var ErrWriter = StringParameter{
//	BaseParameter: BaseParameter{
//		Expression: RFC1123,
//		ShortHelp:  `A writer`,
//		LongHelp: `Where to write an err stream
//		Options:
//		  - stdout
//		  - stdin
//		  - file:<path>
//		  - null
//		`,
//	},
//	Text:   "<err>",
//	Parser: ParseWriter,
//}
