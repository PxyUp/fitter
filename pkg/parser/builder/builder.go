package builder

type Jsonable interface {
	ToJson() string
	IsEmpty() bool
	Raw() interface{}
}

var (
	EMPTY = PureString("")
)
