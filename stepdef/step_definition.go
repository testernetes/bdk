package stepdef

type StepDefinition struct {
	Name     string
	Text     string
	Help     string
	Examples string
	Function any
	StepArg  StepArgument
}
