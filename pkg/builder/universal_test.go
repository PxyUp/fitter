package builder_test

import (
	"github.com/PxyUp/fitter/pkg/builder"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInt(t *testing.T) {
	num := builder.Number(5)
	assert.Equal(t, num.Raw(), builder.ToJsonable(num.Raw()).Raw())
}

func TestFloat(t *testing.T) {
	num := builder.Number(5.555)
	assert.Equal(t, num.Raw(), builder.ToJsonable(num.Raw()).Raw())
}

func TestString(t *testing.T) {
	str := builder.String("5.555", false)
	assert.Equal(t, str.Raw(), builder.ToJsonable(str.Raw()).Raw())
}

func TestBool(t *testing.T) {
	bb := builder.Bool(true)
	assert.Equal(t, bb.Raw(), builder.ToJsonable(bb.Raw()).Raw())
}

func TestNull(t *testing.T) {
	nn := builder.NullValue
	assert.Equal(t, nn.Raw(), builder.ToJsonable(nn.Raw()).Raw())
}

func TestObject(t *testing.T) {
	s := builder.Object(map[string]builder.Interfacable{
		"value": builder.String("asf", false),
	})
	assert.Equal(t, s.Raw(), builder.ToJsonable(s.Raw()).Raw())
}

func TestComplexObject(t *testing.T) {
	s := builder.Object(map[string]builder.Interfacable{
		"test": builder.Array([]builder.Interfacable{
			builder.Number(5), builder.Bool(true), builder.Array([]builder.Interfacable{builder.Number(3)}),
		}),
		"map": builder.Object(map[string]builder.Interfacable{
			"5": builder.Number(5),
			"4": builder.String("fasfasf", false),
			"3": builder.NullValue,
		}),
	})

	assert.Equal(t, s.Raw(), builder.ToJsonable(s.Raw()).Raw())
}

func TestComplexArray(t *testing.T) {
	s := builder.Array([]builder.Interfacable{
		builder.Array([]builder.Interfacable{
			builder.Number(5), builder.Bool(true), builder.Array([]builder.Interfacable{builder.Number(3)}),
		}),
		builder.Object(map[string]builder.Interfacable{
			"5": builder.Number(5),
			"4": builder.String("fasfasf", false),
			"3": builder.NullValue,
		}),
		builder.Number(5),
	})
	assert.Equal(t, s.Raw(), builder.ToJsonable(s.Raw()).Raw())
}

func TestComplexPlain(t *testing.T) {
	value := []byte(`{
  "214": "fsdfa",
  "1233": [1,2,3]
}`)

	ss := builder.Object(map[string]builder.Interfacable{
		"214":  builder.String("fsdfa", false),
		"1233": builder.Array([]builder.Interfacable{builder.Number(1), builder.Number(2), builder.Number(3)}),
	})

	assert.JSONEq(t, string(value), string(builder.ToJsonable(ss.Raw()).Raw()))
}
