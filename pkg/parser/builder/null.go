package builder

type null struct {
}

func Null() *null {
	return &null{}
}

func (n *null) ToJson() string {
	return `null`
}
