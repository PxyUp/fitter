package builder

import "encoding/json"

type nn struct {
}

func (s *nn) ToInterface() interface{} {
	return nil
}

var (
	_ Interfacable = &nn{}
)

func null() *nn {
	return &nn{}
}

func (s *nn) IsEmpty() bool {
	return true
}

func (n *nn) ToJson() string {
	return `null`
}

func (s *nn) Raw() json.RawMessage {
	return json.RawMessage{0x00}
}
