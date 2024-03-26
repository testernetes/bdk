package parameters

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/testernetes/bdk/formatters/utils"
)

type Parameter interface {
	GetShortHelp() string
	GetLongHelp() string
	GetExpression() string
	ConvertsTo(reflect.Kind) bool
	Print() string
}

type StepArgParameter interface {
	Parameter
	IsStepArg() bool
}

type BaseParameter struct {
	Text       string
	ShortHelp  string
	LongHelp   string
	Expression string
	Kinds      []reflect.Kind
	// Add a Supported Kinds array for step validation
}

func (p BaseParameter) ConvertsTo(k reflect.Kind) bool {
	for _, kind := range p.Kinds {
		if kind == k {
			return true
		}
	}
	return false
}

func (p BaseParameter) GetShortHelp() string {
	return p.ShortHelp
}

func (p BaseParameter) GetLongHelp() string {
	return p.LongHelp
}

func (p BaseParameter) GetExpression() string {
	return p.Expression
}

var _ Parameter = (*StringParameter)(nil)

type StringParameter struct {
	BaseParameter
	Parser func(string, reflect.Type) (reflect.Value, error)
}

func (p StringParameter) GetText() string {
	return p.Text
}

func (p StringParameter) Print() string {
	buf := bytes.NewBufferString("")
	sanatized := strings.ReplaceAll(p.Text, "<", "&lt;")
	sanatized = strings.ReplaceAll(p.Text, ">", "&gt;")
	fmt.Fprintf(buf, utils.NewNormalizer(sanatized).Trim().String())
	fmt.Fprintf(buf, utils.Parameter(p.GetShortHelp()))
	fmt.Fprintf(buf, utils.Parameter(p.GetLongHelp()))
	return buf.String()
}

var _ StepArgParameter = (*DocStringParameter)(nil)

type DocStringParameter struct {
	BaseParameter
	Parser func(*messages.DocString, reflect.Type) (reflect.Value, error)
}

func (p DocStringParameter) IsStepArg() bool {
	return true
}

func (p DocStringParameter) Print() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Additional Step Arguments")
	fmt.Fprintf(buf, utils.Parameter(p.GetShortHelp()))
	fmt.Fprintf(buf, utils.Parameter(p.GetLongHelp()))
	return buf.String()
}

var _ StepArgParameter = (*DataTableParameter)(nil)

type DataTableParameter struct {
	BaseParameter
	Parser func(*messages.DataTable, reflect.Type) (reflect.Value, error)
}

func (p DataTableParameter) IsStepArg() bool {
	return true
}

func (p DataTableParameter) Print() string {
	buf := bytes.NewBufferString("")
	header := "**Additional Step Arguments: " + p.ShortHelp + "**"
	fmt.Fprintf(buf, utils.NewNormalizer(header).Trim().Definition().String())
	fmt.Fprintf(buf, "```\n")
	fmt.Fprintf(buf, utils.NewNormalizer(p.GetLongHelp()).TrimAllTabs().String())
	fmt.Fprintf(buf, "```\n")
	return buf.String()
}

var NoStepArg StepArgParameter = &noStepArg{}

type noStepArg struct {
	BaseParameter
}

func (p *noStepArg) IsStepArg() bool {
	return false
}

func (p *noStepArg) Print() string {
	return "No Supported Step Arguments"
}
