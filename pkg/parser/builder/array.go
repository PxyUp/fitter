package builder

type arrayField struct {
	values []Jsonable
}

func Array(items []Jsonable) *arrayField {
	return &arrayField{
		values: items,
	}
}

func (s *arrayField) ToJson() string {
	str := "["

	for i, item := range s.values {
		str += item.ToJson()
		if i != len(s.values) {
			str += ","
		}
	}

	return str + "]"
}
