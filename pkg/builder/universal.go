package builder

import (
	"encoding/json"
	"github.com/tidwall/gjson"
)

func toJson(result gjson.Result) Interfacable {
	if result.IsObject() {
		tmp := make(map[string]Interfacable)

		result.ForEach(func(key, value gjson.Result) bool {
			tmp[key.String()] = toJson(value)
			return true
		})

		return Object(tmp)
	}

	if result.IsArray() {
		var tmp []Interfacable
		result.ForEach(func(key, value gjson.Result) bool {
			tmp = append(tmp, toJson(value))
			return true
		})

		return Array(tmp)
	}

	if result.IsBool() {
		return Bool(result.Bool())
	}

	switch result.Type {
	case gjson.String:
		return String(result.String(), false)
	case gjson.Null:
		return NullValue
	case gjson.Number:
		return Number(result.Num)
	}
	return NullValue
}

func ToJsonable(raw json.RawMessage) Interfacable {
	return toJson(gjson.ParseBytes(raw))
}

func ToJsonableFromString(str string) Interfacable {
	return ToJsonable([]byte(str))
}
