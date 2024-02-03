package builder

import (
	"encoding/json"
	"fmt"
)

type floatField struct {
	value float32
}

type float64Field struct {
	value float64
}

var (
	_ Jsonable = &floatField{}
	_ Jsonable = &float64Field{}
)

func Float(value float32) *floatField {
	return &floatField{
		value: value,
	}
}

func (s *floatField) IsEmpty() bool {
	return false
}

func (s *floatField) ToJson() string {
	return fmt.Sprintf(`%f`, s.value)
}

func (s *floatField) Raw() json.RawMessage {
	return toRaw(s)
}

func Float64(value float64) *float64Field {
	return &float64Field{
		value: value,
	}
}

func (s *float64Field) IsEmpty() bool {
	return false
}

func (s *float64Field) ToJson() string {
	return fmt.Sprintf(`%f`, s.value)
}

func (s *float64Field) Raw() json.RawMessage {
	return toRaw(s.value)
}
