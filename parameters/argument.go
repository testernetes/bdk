package parameters

import "encoding/json"

type StepArgument interface {
	UnmarshalInto(interface{}) error
	json.Marshaler
	json.Unmarshaler
}
