package builder

import "encoding/json"

type null struct {
}

func (s *null) ToInterface() interface{} {
	return nil
}

var (
	_ Interfacable = &null{}
)

func Null() *null {
	return &null{}
}

func (s *null) IsEmpty() bool {
	return true
}

func (n *null) ToJson() string {
	return `null`
}

func (s *null) Raw() json.RawMessage {
	return json.RawMessage{0x00}
}
