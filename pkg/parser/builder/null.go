package builder

type null struct {
}

func Null() *null {
	return &null{}
}

func (s *null) IsEmpty() bool {
	return true
}

func (n *null) ToJson() string {
	return `null`
}
