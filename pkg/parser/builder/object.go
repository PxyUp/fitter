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

	for k, v := range o.kv {
		str += fmt.Sprintf(`"%s": %s`, k, v.ToJson())
	}

	return str + "}"
}
