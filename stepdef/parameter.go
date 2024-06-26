package stepdef

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	messages "github.com/cucumber/messages/go/v21"
)

type StepArgType string

const (
	DocStringStepArgType StepArgType = "DocString"
	DataTableStepArgType StepArgType = "DataTable"
	NoStepArgType        StepArgType = "NoStepArg"
)

type Parameter interface {
	Name() string
	Description() string
	Help() string
	Print() string
}

type StringParameter interface {
	Parameter
	Expression() string
	Parse(context.Context, string, reflect.Type) (reflect.Value, error)
}

type StepArgument interface {
	Parameter
	Parse(context.Context, *messages.Step, reflect.Type) (reflect.Value, error)
	StepArgType() StepArgType
}

var _ StringParameter = (*stringParameter)(nil)

type stringParameter struct {
	name        string
	description string
	help        string
	expression  string
	parser      func(context.Context, string, reflect.Type) (reflect.Value, error)
}

func (p stringParameter) Parse(ctx context.Context, s string, t reflect.Type) (reflect.Value, error) {
	return p.parser(ctx, s, t)
}

func (p stringParameter) Name() string {
	return p.name
}

func (p stringParameter) Expression() string {
	return p.expression
}

func (p stringParameter) Description() string {
	return p.description
}

func (p stringParameter) Help() string {
	return p.help
}

func (p stringParameter) Print() string {
	buf := bytes.NewBufferString("")
	//sanatized := strings.ReplaceAll(p.Text, "<", "&lt;")
	//sanatized = strings.ReplaceAll(p.Text, ">", "&gt;")
	//fmt.Fprintf(buf, utils.NewNormalizer(sanatized).Trim().String())
	//fmt.Fprintf(buf, utils.Parameter(p.GetShortHelp()))
	//fmt.Fprintf(buf, utils.Parameter(p.GetLongHelp()))
	return buf.String()
}

var _ StepArgument = (*DocStringArgument)(nil)

type DocStringArgument struct {
	name        string
	description string
	help        string
	parser      func(context.Context, *messages.DocString, reflect.Type) (reflect.Value, error)
}

func (p DocStringArgument) Name() string {
	return p.name
}

func (p DocStringArgument) Description() string {
	return p.description
}

func (p DocStringArgument) Help() string {
	return p.help
}

func (p DocStringArgument) StepArgType() StepArgType {
	return DocStringStepArgType
}

func (p DocStringArgument) Parse(ctx context.Context, s *messages.Step, t reflect.Type) (reflect.Value, error) {
	if s.DataTable != nil {
		return reflect.Value{}, fmt.Errorf("expected a DocString but found a DataTable")
	}
	return p.parser(ctx, s.DocString, t)
}

func (p DocStringArgument) Print() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Additional Step Arguments")
	//fmt.Fprintf(buf, utils.Parameter(p.GetShortHelp()))
	//fmt.Fprintf(buf, utils.Parameter(p.GetLongHelp()))
	return buf.String()
}

var _ StepArgument = (*dataTableArgument)(nil)

type DataTableParser func(context.Context, *messages.DataTable, reflect.Type) (reflect.Value, error)

func NewDataTableArgument(name, description, help string, parser DataTableParser) StepArgument {
	return &dataTableArgument{
		name:        name,
		description: description,
		help:        help,
		parser:      parser,
	}
}

type dataTableArgument struct {
	name        string
	description string
	help        string
	parser      DataTableParser
}

func (p dataTableArgument) Name() string {
	return p.name
}

func (p dataTableArgument) Description() string {
	return p.description
}

func (p dataTableArgument) Help() string {
	return p.help
}

func (p dataTableArgument) StepArgType() StepArgType {
	return DataTableStepArgType
}

func (p dataTableArgument) Parse(ctx context.Context, s *messages.Step, t reflect.Type) (reflect.Value, error) {
	if s.DocString != nil {
		return reflect.Value{}, fmt.Errorf("expected a DataTable but found a DocString")
	}
	v, err := p.parser(ctx, s.DataTable, t)
	if err != nil {
		return reflect.Value{}, err
	}
	if v.Type() != t {
		return reflect.Value{}, fmt.Errorf("DataTableParser function returned a %s, expected %s", v.Type(), t)
	}
	return v, nil
}

func (p dataTableArgument) Print() string {
	buf := bytes.NewBufferString("")
	//header := "**Additional Step Arguments: " + p.ShortHelp + "**"
	//fmt.Fprintf(buf, utils.NewNormalizer(header).Trim().Definition().String())
	fmt.Fprintf(buf, "```\n")
	//fmt.Fprintf(buf, utils.NewNormalizer(p.GetLongHelp()).TrimAllTabs().String())
	fmt.Fprintf(buf, "```\n")
	return buf.String()
}

var NoStepArg StepArgument = &noStepArg{}

type noStepArg struct{}

func (p noStepArg) Name() string {
	return "No Step Argument"
}

func (p noStepArg) Description() string {
	return "step arguments are not supported"
}

func (p noStepArg) Help() string {
	return ""
}

func (p noStepArg) Parse(ctx context.Context, s *messages.Step, t reflect.Type) (reflect.Value, error) {
	if s.DocString != nil || s.DataTable != nil {
		return reflect.Value{}, fmt.Errorf("step does not support step arguments")
	}
	return reflect.Value{}, nil
}

func (p *noStepArg) StepArgType() StepArgType {
	return NoStepArgType
}

func (p *noStepArg) Print() string {
	return "No Supported Step Arguments"
}
