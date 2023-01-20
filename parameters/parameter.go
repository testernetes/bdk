package parameters

import (
	"reflect"

	"github.com/testernetes/bdk/arguments"
)

type Parameter interface {
	GetShortHelp() string
	GetLongHelp() string
	GetExpression() string
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

var _ Parameter = (*DocStringParameter)(nil)

type DocStringParameter struct {
	BaseParameter
	Parser func(*arguments.DocString, reflect.Type) (reflect.Value, error)
}

var _ Parameter = (*DataTableParameter)(nil)

type DataTableParameter struct {
	BaseParameter
	Parser func(*arguments.DataTable, reflect.Type) (reflect.Value, error)
}
