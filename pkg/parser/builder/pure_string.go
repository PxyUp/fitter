package builder

import "strings"

const (
	EmptyString = `""`
)

type pureStringField struct {
	value string
}

func PureString(value string) *pureStringField {
	value = strings.TrimRight(strings.TrimLeft(value, `"`), `"`)
	if value == "" {
		value = EmptyString
	}
	return &pureStringField{
		value: value,
	}
}

func (s *pureStringField) ToJson() string {
	return s.value
}
