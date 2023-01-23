package parameters

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/testernetes/bdk/arguments"
	"github.com/testernetes/bdk/formatters/utils"
)

type Parameter interface {
	GetShortHelp() string
	GetLongHelp() string
	GetExpression() string
	Print() string
}

type BaseParameter struct {
	ShortHelp  string
	LongHelp   string
	Expression string
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
	Text   string
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

var _ Parameter = (*DocStringParameter)(nil)

type DocStringParameter struct {
	BaseParameter
	Parser func(*arguments.DocString, reflect.Type) (reflect.Value, error)
}

func (p DocStringParameter) Print() string {
	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "Additional Step Arguments")
	fmt.Fprintf(buf, utils.Parameter(p.GetShortHelp()))
	fmt.Fprintf(buf, utils.Parameter(p.GetLongHelp()))
	return buf.String()
}

var _ Parameter = (*DataTableParameter)(nil)

type DataTableParameter struct {
	BaseParameter
	Parser func(*arguments.DataTable, reflect.Type) (reflect.Value, error)
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
