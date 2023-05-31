package parameters

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"

	"github.com/testernetes/bdk/arguments"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		| dry run       | boolean |
		| field manager | string  |`,
	},
	Parser: func(table *arguments.DataTable, targetType reflect.Type) (reflect.Value, error) {
		opts := []client.CreateOption{}
		if table == nil {
			return reflect.ValueOf(opts), nil
		}

		for _, row := range table.Rows {
			if len(row.Cells) < 2 {
				continue
			}
			opt := row.Cells[0].Value
			val := row.Cells[1].Value
			switch opt {
			case "dry run":
				if val == "true" {
					opts = append(opts, client.DryRunAll)
				}
			case "field owner":
				opts = append(opts, client.FieldOwner(val))
			default:
				return reflect.Value{}, fmt.Errorf("invalid create option: %s", opt)
			}
		}
		return reflect.ValueOf(opts), nil
	},
}

var DeleteOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Slice},
		ShortHelp: `(optional) A table of additional client delete options.`,
		LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#DeleteOptions

		Delete Options:
		| dry run              | boolean                        |
		| grace period seconds | number                         |
		| propagation policy   | (Orphan|Background|Foreground) |`,
	},
	Parser: func(table *arguments.DataTable, targetType reflect.Type) (reflect.Value, error) {
		opts := []client.DeleteOption{}
		if table == nil {
			return reflect.ValueOf(opts), nil
		}
		for _, row := range table.Rows {
			if len(row.Cells) < 2 {
				continue
			}
			opt := row.Cells[0].Value
			val := row.Cells[1].Value
			switch opt {
			case "dry run":
				if val == "true" {
					opts = append(opts, client.DryRunAll)
				}
			case "grace period seconds":
				v, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					panic(err)
				}
				opts = append(opts, client.GracePeriodSeconds(v))
			case "preconditions":
				panic("preconditions not yet supported")
			case "propagation policy":
				opts = append(opts, client.PropagationPolicy(val))
			default:
				panic(fmt.Sprintf("invalid delete option %s", opt))
			}
		}
		return reflect.ValueOf(opts), nil
	},
}

var PatchOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Slice},
		ShortHelp: `A table of additional client patch options.`,
		LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#PatchOptions

		Patch Options:
		| patch         | string (required) |
		| force         | boolean           |
		| dry run       | boolean           |
		| field manager | string            |`,
	},
	Parser: func(table *arguments.DataTable, targetType reflect.Type) (reflect.Value, error) {
		opts := []interface{}{}
		if table == nil {
			return reflect.ValueOf(opts), nil
		}
		for _, row := range table.Rows {
			if len(row.Cells) < 2 {
				continue
			}
			opt := row.Cells[0].Value
			val := row.Cells[1].Value
			switch opt {
			case "dry run":
				if val == "true" {
					opts = append(opts, client.DryRunAll)
				}
			case "field owner":
				opts = append(opts, client.FieldOwner(val))
			case "force":
				if val == "true" {
					opts = append(opts, client.ForceOwnership)
				}
			case "patch":
				patch, err := yaml.YAMLToJSON([]byte(val))
				if err != nil {
					panic(err)
				}
				opts = append(opts, client.RawPatch(types.StrategicMergePatchType, patch))
			}
		}
		return reflect.ValueOf(opts), nil
	},
}

var PodLogOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Ptr},
		ShortHelp: `(optional) A table of additional client pod log options.`,
		LongHelp: `https://pkg.go.dev/k8s.io/api/core/v1#PodLogOptions

		Pod Log Options:
		| container     | string  |
		| follow        | boolean |
		| previous      | boolean |
		| since seconds | number  |
		| timestamps    | boolean |
		| tail lines    | number  |
		| limit bytes   | number  |`,
	},
	Parser: func(table *arguments.DataTable, targetType reflect.Type) (reflect.Value, error) {
		// TODO Add out and err Writer to these options
		opts := &corev1.PodLogOptions{}
		if table == nil {
			return reflect.ValueOf(opts), nil
		}
		for _, row := range table.Rows {
			if len(row.Cells) < 2 {
				continue
			}
			opt := row.Cells[0].Value
			val := row.Cells[1].Value
			switch opt {
			case "container":
				opts.Container = val
			case "follow":
				opts.Follow = val == "true"
			case "previous":
				opts.Previous = val == "true"
			case "since seconds":
				v, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					panic(err)
				}
				opts.SinceSeconds = &v
			case "since time":
				panic("since time not yet implemented")
			case "timestamps":
				opts.Timestamps = val == "true"
			case "tail lines":
				v, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					panic(err)
				}
				opts.TailLines = &v
			case "limit bytes":
				v, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					panic(err)
				}
				opts.LimitBytes = &v
			}
		}
		return reflect.ValueOf(opts), nil
	},
}

var ProxyGetOptions = DataTableParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Map},
		ShortHelp: `(optional) A freeform table of additional query parameters to send with the request.`,
		LongHelp: `

		ProxyGet Options:
		| string | string |`,
	},
	Parser: func(table *arguments.DataTable, targetType reflect.Type) (reflect.Value, error) {
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

var Filename = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.Ptr},
		Expression: Anything,
		ShortHelp:  `Path to a Kubernetes manifest.`,
		LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/

		Can be yaml or json depending on the content type.`,
	},
	Text: "<filename>",
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
}

var Manifest = DocStringParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.Ptr},
		ShortHelp: `A Kubernetes manifest.`,
		LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/

		Can be yaml or json depending on the content type.`,
	},
	Parser: func(docString *arguments.DocString, targetType reflect.Type) (reflect.Value, error) {
		u := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(docString.Content), u)
		if err != nil {
			return reflect.ValueOf(u), err
		}

		var (
			ErrNoAPIVersion = errors.New("Provided test case resource has an empty API Version")
			ErrNoKind       = errors.New("Provided test case resource has an empty Kind")
			ErrNoName       = errors.New("Provided test case resource has an empty Name")
		)

		if u.GetAPIVersion() == "" {
			return reflect.ValueOf(u), ErrNoAPIVersion
		}
		if u.GetKind() == "" {
			return reflect.ValueOf(u), ErrNoKind
		}
		if u.GetName() == "" {
			return reflect.ValueOf(u), ErrNoName
		}
		return reflect.ValueOf(u), nil
	},
}

var Script = DocStringParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.String},
		ShortHelp: `A script.`,
		LongHelp: `The script will run in the specified shell or if none is specified /bin/sh.
		Its outputs will be captured and can be asserted against in future steps.`,
	},
	Parser: DocStringParseString,
}

var MultilineText = DocStringParameter{
	BaseParameter: BaseParameter{
		Kinds:     []reflect.Kind{reflect.String},
		ShortHelp: `A freeform DocString.`,
		LongHelp:  `Any multiline text.`,
	},
	Parser: DocStringParseString,
}

var AsyncAssertionPhrase = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: AsyncType,
		ShortHelp:  `An assertion that state should be achieved or maintained.`,
		LongHelp: fmt.Sprintf(`
		Eventually assertions can be made with: %q
		Consistently assertions can be made with: %q`, EventuallyPhrases, ConsistentlyPhrases),
	},
	Text:   "<assertion>",
	Parser: ParseString,
}

var Reference = StringParameter{
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
	Text:   "<reference>",
	Parser: ParseString,
}

var Command = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: DoubleQuoted,
		ShortHelp:  `The command to execute in the container.`,
		LongHelp: `https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/

		The command will run in a shell on the container and its outputs will be captured and can
		be asserted against in future steps.`,
	},
	Text:   "<command>",
	Parser: ParseString,
}

var Container = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: RFC1123,
		ShortHelp:  `The container name.`,
		LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
		
		The reference must a name that can be used as a DNS subdomain name as defined in RFC 1123.
		This is the same Kubernetes requirement for names, i.e. lowercase alphanumeric characters.`,
	},
	Text:   "<container>",
	Parser: ParseString,
}

var Duration = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.Int64},
		Expression: exprDuration,
		ShortHelp:  `Duration from when the step starts.`,
		LongHelp: `https://pkg.go.dev/time#ParseDuration
		
		A duration is a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5m" or "2h45m".
		Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`,
	},
	Text: "<duration>",
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
}

var JSONPath = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: SingleQuoted,
		ShortHelp:  `A JSON Path to a field in the referenced resource.`,
		LongHelp: `https://kubernetes.io/docs/reference/kubectl/jsonpath/
		
		e.g. '{.metadata.name}'.`,
	},
	Text:   "<jsonpath>",
	Parser: ParseString,
}

var Matcher = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: Anything,
		ShortHelp:  `Used in conjunction with an assertion to assert that the actual matches the expected.`,
		LongHelp:   `To list available matchers run 'bdk matchers'.`,
	},
	Text:   "<matcher>",
	Parser: ParseString,
}

var Text = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: Anything,
		ShortHelp:  `A freeform amount of text.`,
		LongHelp:   `This will match anything.`,
	},
	Text:   "<text>",
	Parser: ParseString,
}

var Number = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64},
		Expression: exprNumber,
		ShortHelp:  `A number.`,
		LongHelp:   `Can be decimal.`,
	},
	Text:   "<number>",
	Parser: ParseNumber,
}

var Array = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.Slice},
		Expression: exprArray,
		ShortHelp:  `A set of values.`,
		LongHelp:   `Must be space delimited.`,
	},
	Text:   "<array>",
	Parser: ParseArray,
}

var Comparator = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: exprComparator,
		ShortHelp:  `a numeric comparator.`,
		LongHelp:   `One of ==, <, >, <=, >=.`,
	},
	Text:   "<comparator>",
	Parser: ParseString,
}

var ShouldOrShouldNot = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: exprShouldOrShouldNot,
		ShortHelp:  `A positive or negative assertion.`,
		LongHelp:   `"to" can also be used instead of "should".`,
	},
	Text:   "(should|should not)",
	Parser: ParseString,
}

var Port = StringParameter{
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
}

var URLPath = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: exprURLPath,
		ShortHelp:  `The path of a URL.`,
		LongHelp:   `Anything that comes after port.`,
	},
	Text:   "<path>",
	Parser: ParseString,
}

var URLScheme = StringParameter{
	BaseParameter: BaseParameter{
		Kinds:      []reflect.Kind{reflect.String},
		Expression: exprURLScheme,
		ShortHelp:  `The scheme of a URL.`,
		LongHelp:   `Must be either http or https.`,
	},
	Text:   "<scheme>",
	Parser: ParseString,
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
