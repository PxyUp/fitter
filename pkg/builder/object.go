package builder

import (
	"encoding/json"
	"fmt"
)

type objectField struct {
	kv map[string]Interfacable
}

func (s *objectField) ToInterface() interface{} {
	kv := make(map[string]interface{})
	for k, v := range s.kv {
		kv[k] = v.ToInterface()
	}

	return kv
}

var (
	_ Interfacable = &objectField{}
)

func Object(values map[string]Interfacable) *objectField {
	return &objectField{
		kv: values,
	}
}

func (s *objectField) IsEmpty() bool {
	if len(s.kv) == 0 {
		return true
	}

	for _, v := range s.kv {
		if !v.IsEmpty() {
			return false
		}
	}
	return true
}

func (o *objectField) ToJson() string {
	str := "{"
	index := 0
	for k, v := range o.kv {
		str += fmt.Sprintf(`"%s": %s`, k, v.ToJson())
		if index != len(o.kv)-1 {
			str += ","
		}
		index += 1
	}

	return str + "}"
}

func (o *objectField) Raw() json.RawMessage {
	kv := make(map[string]interface{})
	for k, v := range o.kv {
		kv[k] = v.ToInterface()
	}
	return toRaw(kv)
}
