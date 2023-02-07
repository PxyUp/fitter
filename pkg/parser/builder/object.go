package builder

import "fmt"

type objectField struct {
	kv map[string]Jsonable
}

func Object(values map[string]Jsonable) *objectField {
	return &objectField{
		kv: values,
	}
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
