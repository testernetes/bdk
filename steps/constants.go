package steps

const (
	NamedObj           = `([a-z0-9]+[-a-z0-9]*[a-z0-9])`
	DoubleQuoted       = `"([^"\\]*(?:\\.[^"\\]*)*)"`
	SingleQuoted       = `'([^'\\]*(?:\\.[^'\\]*)*)'`
	OneOrMultipleWords = `([\w+\s*]+)`
	AsyncType          = OneOrMultipleWords
	Duration           = `([\d+\w{1,2}]+)`
	ShouldOrShouldNot  = `(?:should|to)( not)?`
	Anything           = `(.*)`
	Matcher            = Anything
	URLPath            = `([-a-zA-Z0-9()!@:%_\+.~#?&\/\/=]*)`
	URLScheme          = `(http|https)`
	Port               = `(\d{1,5})`
)

//        Unspecified,
//        Context,
//        Action,
//        Outcome,
//        Conjunction,
//        Unknown

const (
	ErrNoResource string = "No resource called %s was registered in a previous step"
)
