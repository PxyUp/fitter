package builder

import (
	"encoding/json"
	"fmt"
)

type number struct {
	value float64
}

func (s *number) ToInterface() interface{} {
	return s.value
}

var (
	_ Interfacable = &number{}
)

func Number(value float64) *number {
	return &number{
		value: value,
	}
}

func isIntegral(val float64) bool {
	return val == float64(int(val))
}

func (s *number) IsEmpty() bool {
	return false
}

func (s *number) ToJson() string {
	if isIntegral(s.value) {
		return fmt.Sprintf("%d", int(s.value))
	}
	return fmt.Sprintf("%v", s.value)
}

func (s *number) Raw() json.RawMessage {
	return toRaw(s.value)
}
