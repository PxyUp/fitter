package builder

import "strings"

type pureStringField struct {
	value string
}

func PureString(value string) *pureStringField {
	value = strings.TrimRight(strings.TrimLeft(value, `"`), `"`)
	if value == "" {
		value = `""`
	}
	return &pureStringField{
		value: value,
	}
}

func (s *pureStringField) ToJson() string {
	return s.value
}
