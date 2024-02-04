package builder

import (
	"encoding/json"
	"fmt"
)

type boolField struct {
	value bool
}

func (s *boolField) ToInterface() interface{} {
	return s.value
}

var (
	_ Interfacable = &boolField{}
)

func Bool(value bool) *boolField {
	return &boolField{
		value: value,
	}
}

func (s *boolField) IsEmpty() bool {
	return false
}

func (s *boolField) ToJson() string {
	return fmt.Sprintf(`%v`, s.value)
}

func (s *boolField) Raw() json.RawMessage {
	return toRaw(s.value)
}
