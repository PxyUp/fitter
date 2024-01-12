package builder

import (
	"golang.org/x/net/html"
	"strconv"
	"strings"
)

type stringField struct {
	value string
}

var (
	_ Jsonable = &stringField{}
)

func String(value string) *stringField {
	return &stringField{
		value: strings.TrimSpace(value),
	}
}

func (s *stringField) IsEmpty() bool {
	return len(s.value) == 0
}

func (s *stringField) ToJson() string {
	return strconv.Quote(html.EscapeString(s.value))
}

func (s *stringField) Raw() interface{} {
	return s.value
}
