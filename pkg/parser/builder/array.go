package builder

type arrayField struct {
	values []Jsonable
}

func Array(items []Jsonable) *arrayField {
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
		str += item.ToJson()
		if i != len(s.values)-1 {
			str += ","
		}
	}

	return str + "]"
}
