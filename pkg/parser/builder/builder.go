package builder

type Jsonable interface {
	ToJson() string
}

var (
	EMPTY = PureString("")
)
