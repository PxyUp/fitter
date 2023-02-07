package builder

import "fmt"

type boolField struct {
	value bool
}

func Bool(value bool) *boolField {
	return &boolField{
		value: value,
	}
}

func (s *boolField) ToJson() string {
	return fmt.Sprintf(`%v`, s.value)
}
