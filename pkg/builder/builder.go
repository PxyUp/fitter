package builder

import "encoding/json"

type Jsonable interface {
	ToJson() string
	IsEmpty() bool
	Raw() json.RawMessage
}

var (
	EMPTY = PureString("")

	NullValue = Null()
)

func toRaw(vv any) json.RawMessage {
	bb, err := json.Marshal(vv)
	if err != nil {
		return NullValue.Raw()
	}

	return bb
}
