package parameters

import (
	"fmt"
	"reflect"

	"github.com/testernetes/bdk/arguments"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

//        Unspecified,
//        Context,
//        Action,
//        Outcome,
//        Conjunction,
//        Unknown

type Parameter struct {
	Text       string
	ShortHelp  string
	LongHelp   string
	Expression string
	Converter  func(interface{}) (reflect.Value, error)
}

var CreateOptions = Parameter{
	ShortHelp: `(optional) A table of additional client create options.`,
	LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#CreateOptions

		Create Options:
		| dry run       | boolean |
		| field manager | string  |`,
	Converter: func(arg interface{}) (reflect.Value, error) {
		opts := []client.CreateOption{}
		table, ok := arg.(*arguments.DataTable)
		if table == nil || !ok {
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

var DeleteOptions = Parameter{
	ShortHelp: `(optional) A table of additional client delete options.`,
	LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#DeleteOptions

		Delete Options:
		| dry run              | boolean                        |
		| grace period seconds | number                         |
		| propagation policy   | (Orphan|Background|Foreground) |`,
}

var PatchOptions = Parameter{
	ShortHelp: `A table of additional client patch options.`,
	LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#PatchOptions

		Patch Options:
		| patch         | string (required) |
		| force         | boolean           |
		| dry run       | boolean           |
		| field manager | string            |`,
}

var PodLogOptions = Parameter{
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
}

var ProxyGetOptions = Parameter{
	ShortHelp: `(optional) A freeform table of additional query parameters to send with the request.`,
	LongHelp: `

		ProxyGet Options:
		| string | string |`,
}

var Manifest = Parameter{
	ShortHelp: `A Kubernetes manifest.`,
	LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/

		Can be yaml or json depending on the content type.`,
}

var Script = Parameter{
	ShortHelp: `A script.`,
	LongHelp: `The script will run in the specified shell or if none is specified /bin/sh.
		Its outputs will be captured and can be asserted against in future steps.`,
}

var MultilineText = Parameter{
	ShortHelp: `A freeform DocString.`,
	LongHelp:  `Any multiline text.`,
}

var AsyncAssertion = Parameter{
	Text:       "<assertion>",
	Expression: AsyncType,
	ShortHelp:  `An assertion that state should be achieved or maintained.`,
	LongHelp: fmt.Sprintf(`
		Eventually assertions can be made with: %q
		Consistently assertions can be made with: %q`, EventuallyPhrases, ConsistentlyPhrases),
}

var Reference = Parameter{
	Text:       "<reference>",
	Expression: RFC1123,
	ShortHelp:  `A short hand name for a resource.`,
	LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
		
		A resource can be assigned to this reference in a Context step and later referred to in an
		Action or Outcome step.
		
		The reference must a name that can be used as a DNS subdomain name as defined in RFC 1123.
		This is the same Kubernetes requirement for names, i.e. lowercase alphanumeric characters.`,
}

var Command = Parameter{
	Text:       "<command>",
	Expression: DoubleQuoted,
	ShortHelp:  `The command to execute in the container.`,
	LongHelp: `https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/

		The command will run in a shell on the container and its outputs will be captured and can
		be asserted against in future steps.`,
}

var Container = Parameter{
	Text:       "<container>",
	Expression: RFC1123,
	ShortHelp:  `The container name.`,
	LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
		
		The reference must a name that can be used as a DNS subdomain name as defined in RFC 1123.
		This is the same Kubernetes requirement for names, i.e. lowercase alphanumeric characters.`,
}

var Duration = Parameter{
	Text:       "<duration>",
	Expression: exprDuration,
	ShortHelp:  `Duration from when the step starts.`,
	LongHelp: `https://pkg.go.dev/time#ParseDuration
		
		A duration is a sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "1.5m" or "2h45m".
		Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`,
}

var JSONPath = Parameter{
	Text:       "<jsonpath>",
	Expression: SingleQuoted,
	ShortHelp:  `A JSON Path to a field in the referenced resource.`,
	LongHelp: `https://kubernetes.io/docs/reference/kubectl/jsonpath/
		
		e.g. '{.metadata.name}'.`,
}

var Matcher = Parameter{
	Text:       "<matcher>",
	Expression: Anything,
	ShortHelp:  `Used in conjunction with an assertion to assert that the actual matches the expected.`,
	LongHelp:   `To list available matchers run 'bdk matchers'.`,
}

var Text = Parameter{
	Text:       "<text>",
	Expression: Anything,
	ShortHelp:  `A freeform amount of text.`,
	LongHelp:   `This will match anything.`,
}

var Number = Parameter{
	Text:       "<number>",
	Expression: exprNumber,
	ShortHelp:  `A number.`,
	LongHelp:   `Can be decimal.`,
}

var Array = Parameter{
	Text:       "<array>",
	Expression: exprArray,
	ShortHelp:  `A set of values.`,
	LongHelp:   `Must be space delimited.`,
}

var Comparator = Parameter{
	Text:       "<comparator>",
	Expression: exprComparator,
	ShortHelp:  `a numeric comparator.`,
	LongHelp:   `One of ==, <, >, <=, >=.`,
}

var ShouldOrShouldNot = Parameter{
	Text:       "(should|should not)",
	Expression: exprShouldOrShouldNot,
	ShortHelp:  `A positive or negative assertion.`,
	LongHelp:   `"to" can also be used instead of "should".`,
}

var Port = Parameter{
	Text:       "<port>",
	Expression: exprPort,
	ShortHelp:  `Port.`,
	LongHelp:   `The port number to request. Acceptable range is 0 - 65536.`,
}

var URLPath = Parameter{
	Text:       "<path>",
	Expression: exprURLPath,
	ShortHelp:  `The path of a URL.`,
	LongHelp:   `Anything that comes after port.`,
}

var URLScheme = Parameter{
	Text:       "<scheme>",
	Expression: exprURLScheme,
	ShortHelp:  `The scheme of a URL.`,
	LongHelp:   `Must be either http or https.`,
}
