package builder

import "fmt"

type floatField struct {
	value float32
}

func Float(value float32) *floatField {
	return &floatField{
		value: value,
	}
}

func (s *floatField) ToJson() string {
	return fmt.Sprintf(`%f`, s.value)
}
