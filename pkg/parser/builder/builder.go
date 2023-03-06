package builder

type Jsonable interface {
	ToJson() string
	IsEmpty() bool
}

var (
	EMPTY = PureString("")
)
