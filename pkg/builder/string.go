package builder

import (
	"encoding/json"
	"golang.org/x/net/html"
	"strconv"
	"strings"
)

type stringField struct {
	value string
}

var (
	_ Jsonable = &stringField{}
	r          = strings.NewReplacer("\n", "", "\r", "", "\t", "")
)

func String(value string, trim ...bool) *stringField {
	if len(trim) > 0 {
		if !trim[0] {
			return &stringField{
				value: value,
			}
		}
	}

	return &stringField{
		value: r.Replace(strings.TrimSpace(value)),
	}
}

func (s *stringField) IsEmpty() bool {
	return len(s.value) == 0
}

func (s *stringField) ToJson() string {
	return strconv.Quote(html.EscapeString(s.value))
}

func (s *stringField) Raw() json.RawMessage {
	return toRaw(s.value)
}
