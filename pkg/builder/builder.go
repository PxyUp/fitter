package builder

import (
	"encoding/json"
)

type Interfacable interface {
	Jsonable

	ToInterface() interface{}
}

type Jsonable interface {
	ToJson() string
	IsEmpty() bool
	Raw() json.RawMessage
}

var (
	EMPTY = PureString("")

	NullValue = null()
)

func toRaw(vv any) json.RawMessage {
	var bb json.RawMessage
	bb, err := json.Marshal(vv)
	if err != nil {
		return NullValue.Raw()
	}

	return bb
}
