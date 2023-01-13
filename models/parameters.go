package models

const (
	RFC1123            = `([a-z0-9]+[-a-z0-9]*[a-z0-9])`
	DoubleQuoted       = `"([^"\\]*(?:\\.[^"\\]*)*)"`
	SingleQuoted       = `'([^'\\]*(?:\\.[^'\\]*)*)'`
	OneOrMultipleWords = `([\w+\s*]+)`
	AsyncType          = OneOrMultipleWords
	Duration           = `([\d+\w{1,2}]+)`
	ShouldOrShouldNot  = `(?:should|to)( not)?`
	Anything           = `(.*)`
	Matcher            = Anything
	exprURLPath        = `([-a-zA-Z0-9()!@:%_\+.~#?&\/\/=]*)`
	exprURLScheme      = `(http|https)`
	exprPort           = `(\d{1,5})`
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
}

var CreateOptions = Parameter{
	ShortHelp: `(optional) A table of additional client create options`,
	LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#CreateOptions

		Create Options:
		| dry run       | (true|false)          |
		| field manager | name of field manager |`,
}

var DeleteOptions = Parameter{
	ShortHelp: `(optional) A table of additional client delete options`,
	LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#DeleteOptions

		Delete Options:
		| dry run              | (true|false)                   |
		| grace period seconds | number                         |
		| propagation policy   | (Orphan|Background|Foreground) |`,
}

var PatchOptions = Parameter{
	ShortHelp: `A table of additional client patch options`,
	LongHelp: `https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#PatchOptions

		Patch Options:
		| patch         | patch (required)      |
		| force         | (true|false)          |
		| dry run       | (true|false)          |
		| field manager | name of field manager |`,
}

var ProxyGetOptions = Parameter{
	ShortHelp: `(optional) A freeform table of additional query parameters to send with the request`,
	LongHelp: `

		ProxyGet Options:
		| anything | anything |`,
}

var Manifest = Parameter{
	ShortHelp: `A Kubernetes manifest.`,
	LongHelp: `https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/

		Can be yaml or json depending on the content type`,
}

var Script = Parameter{
	ShortHelp: `A script.`,
	LongHelp: `The script will run in the specified shell or if none is specified /bin/sh.
		Its outputs will be captured and can be asserted against in future steps.`,
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
