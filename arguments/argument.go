package arguments

import "encoding/json"

type Argument interface {
	UnmarshalInto(interface{}) error
	json.Marshaler
	json.Unmarshaler
}
