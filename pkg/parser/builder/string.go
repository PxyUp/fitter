package builder

import (
	"fmt"
	"strings"
)

type stringField struct {
	value string
}

func String(value string) *stringField {
	return &stringField{
		value: strings.TrimSpace(value),
	}
}

func (s *stringField) IsEmpty() bool {
	return len(s.value) == 0
}

func (s *stringField) ToJson() string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(s.value, `"`, `\"`))
}
