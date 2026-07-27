package builder

import "encoding/json"

// omitted marks a value rejected by a condition: the parent object/array must
// drop it entirely instead of rendering null
type omitted struct {
}

var (
	_ Interfacable = &omitted{}

	OmitValue = &omitted{}
)

func IsOmitted(value Interfacable) bool {
	return value == Interfacable(OmitValue)
}

func (s *omitted) ToInterface() interface{} {
	return nil
}

func (s *omitted) IsEmpty() bool {
	return true
}

func (s *omitted) ToJson() string {
	return `null`
}

func (s *omitted) Raw() json.RawMessage {
	return NullValue.Raw()
}
