package builder

import "fmt"

type intField struct {
	value int
}

func Int(value int) *intField {
	return &intField{
		value: value,
	}
}

func (s *intField) ToJson() string {
	return fmt.Sprintf(`%d`, s.value)
}
