package builder

import "fmt"

type stringField struct {
	value string
}

func String(value string) *stringField {
	return &stringField{
		value: value,
	}
}

func (s *stringField) ToJson() string {
	return fmt.Sprintf(`"%s"`, s.value)
}
