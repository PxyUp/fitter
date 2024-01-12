package builder

type null struct {
}

var (
	_ Jsonable = &null{}
)

func Null() *null {
	return &null{}
}

func (s *null) IsEmpty() bool {
	return true
}

func (n *null) ToJson() string {
	return `null`
}

func (s *null) Raw() interface{} {
	return nil
}
