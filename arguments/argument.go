package arguments

type Argument interface {
	UnmarshalInto(interface{}) error
}
