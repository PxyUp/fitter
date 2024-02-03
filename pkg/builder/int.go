package builder

import (
	"encoding/json"
	"fmt"
)

type intField struct {
	value int
}

type int64Field struct {
	value int64
}

var (
	_ Jsonable = &intField{}
	_ Jsonable = &int64Field{}
)

func Int(value int) *intField {
	return &intField{
		value: value,
	}
}

func (s *intField) IsEmpty() bool {
	return false
}

func (s *intField) ToJson() string {
	return fmt.Sprintf(`%d`, s.value)
}

func (s *intField) Raw() json.RawMessage {
	return toRaw(s.value)
}

func Int64(value int64) *int64Field {
	return &int64Field{
		value: value,
	}
}

func (s *int64Field) IsEmpty() bool {
	return false
}

func (s *int64Field) ToJson() string {
	return fmt.Sprintf(`%d`, s.value)
}

func (s *int64Field) Raw() json.RawMessage {
	return toRaw(s.value)
}
