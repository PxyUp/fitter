package builder

import "encoding/json"

type arrayField struct {
	values []Interfacable
}

func (s *arrayField) ToInterface() interface{} {
	res := make([]interface{}, len(s.values))

	for i, item := range s.values {
		if item == nil {
			res[i] = NullValue.ToInterface()
		} else {
			res[i] = item.ToInterface()
		}
	}

	return res
}

var (
	_ Interfacable = &arrayField{}
)

func Array(items []Interfacable) *arrayField {
	return &arrayField{
		values: items,
	}
}

func (s *arrayField) IsEmpty() bool {
	if len(s.values) == 0 {
		return true
	}

	for _, v := range s.values {
		if !v.IsEmpty() {
			return false
		}
	}

	return true
}

func (s *arrayField) ToJson() string {
	str := "["

	for i, item := range s.values {
		if item == nil {
			str += NullValue.ToJson()
		} else {
			str += item.ToJson()
		}

		if i != len(s.values)-1 {
			str += ","
		}
	}

	return str + "]"
}

func (s *arrayField) Raw() json.RawMessage {
	res := make([]interface{}, len(s.values))

	for i, item := range s.values {
		if item == nil {
			res[i] = NullValue.ToInterface()
		} else {
			res[i] = item.ToInterface()
		}
	}

	return toRaw(res)
}
